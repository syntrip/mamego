/*
 * mamego - terminal mame frontend
 * (c) syntrip sistemas
 * ver 0.5.2 alfa
 * Wed May  6 21:44:52 ART 2015
 *
 */

package main

import (
	"bufio"
	ui "github.com/gizak/termui"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// --- Config file and defaults values
const default_rompath string = "/usr/local/share/games/mame/roms/"
const config_file string = ".mamegorc"
const log_file string = ".mamego.log"

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Search slices for first ocurrence
func firstO(s string, l1 []string, l2 []string) int {
	p := 0
	for p = 0; p < len(l1); p++ {
		if strings.Index(l1[p], s) == 0 || strings.Index(l2[p], s) == 0 {
			break
		}
	}
	if p > len(l1) {
		return -1
	} else {
		return p
	}
}

// Check and select color
func select_color(color string) ui.Attribute {

	c := ui.ColorDefault
	switch color {
	// default, black, red, green, yellow, blue, magenta, cyan, white
	case "black":
		c = ui.ColorBlack
	case "red":
		c = ui.ColorRed
	case "green":
		c = ui.ColorGreen
	case "yellow":
		c = ui.ColorYellow
	case "blue":
		c = ui.ColorBlue
	case "magenta":
		c = ui.ColorMagenta
	case "cyan":
		c = ui.ColorCyan
	case "white":
		c = ui.ColorWhite
	}
	return c

}

func main() {

	config := make(map[string]string)

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	home_dir := usr.HomeDir + "/"

	f, err := os.OpenFile(home_dir+log_file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("error opening file: " + err.Error())
	}
	defer f.Close()

	log.SetOutput(f)

	// config BGN
	config["rompath"] = default_rompath
	//config["item_selected_fg_color"] = "black"
	config["item_selected_bg_color"] = "green"

	file, err := os.Open(home_dir + config_file)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.SplitN(scanner.Text(), "#", 2)
		if len(line[0]) > 0 {
			command := strings.Split(line[0], " ")
			if len(command[0]) > 0 {
				switch command[0] {
				case "rompath":
					config["rompath"] = command[1]
				case "item_fg_color":
					config["item_fg_color"] = command[1]
				case "item_bg_color":
					config["item_bg_color"] = command[1]
				case "item_selected_fg_color":
					config["item_selected_fg_color"] = command[1]
				case "item_selected_bg_color":
					config["item_selected_bg_color"] = command[1]
				default:
					log.Println("WARNING: bad option in " + config_file + "\n\r" + strings.Join(line, " ") + "\r")
				}
			}
		}
	}
	// config END

	ui.UseTheme("default")
	//ui.UseTheme("helloworld")

	// --- fulllist BGN
	cmd := exec.Command("mame", "-ll")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// read command's stdout line by line
	in := bufio.NewScanner(stdout)

	fulllist := make(map[string]string)
	for in.Scan() {
		rom := strings.SplitN(in.Text(), " ", 2)
		fulllist[rom[0]] = strings.Trim(rom[1], " \"")
	}
	if err := in.Err(); err != nil {
		log.Printf("error: %s", err)
	}
	// --- fulllist END

	roms := []string{}
	desc := []string{}
	strs := []string{}

	if ok, err := exists(config["rompath"]); !ok {
		log.Println("ERROR! Directory [" + config["rompath"] + "] doesn't exists\r")
		if err != nil {
			log.Fatal(err)
		}
	}
	files, _ := ioutil.ReadDir(config["rompath"])
	for _, f := range files {
		if !f.IsDir() {
			rn := strings.Split(f.Name(), ".")[0]
			if rd, ok := fulllist[rn]; ok {
				roms = append(roms, rn)
				desc = append(desc, strings.ToLower(rd))
				sep := strings.Repeat(" ", 15-len(rn))
				strs = append(strs, rn+sep+rd)
			}
		}
	}
	buf_init := 0
	buf_size := ui.TermHeight() - 2
	last_action_search := false

	ls := ui.NewPointedList()
	ls.Items = strs[buf_init:]
	ls.ItemFgColor = select_color(config["item_fg_color"])
	ls.ItemBgColor = select_color(config["item_bg_color"])
	ls.ItemSelectedFgColor = select_color(config["item_selected_fg_color"])
	ls.ItemSelectedBgColor = select_color(config["item_selected_bg_color"])
	ls.Border.Label = strconv.Itoa(len(roms)) + " " + "Roms"
	ls.Height = ui.TermHeight()
	ls.Width = ui.TermWidth()
	ls.ItemPointed = 0

	ui.Render(ls)

	// event handler...
	filter := ""
	evt := ui.EventCh()
	for {
		select {
		case e := <-evt:
			if e.Type == ui.EventKey {
				if e.Key == ui.KeyEsc {
					return
				}
				if (unicode.IsPrint(e.Ch) || e.Key == ui.KeySpace) && !(e.Ch == '+' || e.Ch == '-') { // IsGraphic tabs and other white spaces
					if !last_action_search {
						filter = ""
					}
					last_action_search = true
					if e.Key == ui.KeySpace {
						filter += " "
					} else {
						filter += string(e.Ch)
					}
					ls.Border.Label = strconv.Itoa(len(roms)) + " " + "Roms" + " [" + filter + "]"
					pos := firstO(strings.ToLower(filter), roms, desc)
					if pos >= 0 && pos < len(roms) {
						if pos > buf_init && pos < buf_init+buf_size {
							ls.ItemPointed = pos - buf_init
						} else if pos < len(roms)-buf_size {
							ls.ItemPointed = 0
							buf_init = pos
							ls.Items = strs[buf_init:]
						} else {
							buf_init = len(roms) - buf_size
							ls.ItemPointed = pos - buf_init
							ls.Items = strs[buf_init:]
						}
					}
				} else if e.Key == ui.KeyBackspace || e.Key == ui.KeyBackspace2 {
					if fl := len(filter); fl > 0 {
						filter = filter[0 : fl-1]
						ls.Border.Label = strconv.Itoa(len(roms)) + " " + "Roms"
						if fl > 1 {
							ls.Border.Label += " [" + filter + "]"
						}
					}
				} else {
					last_action_search = false
					ls.Border.Label = strconv.Itoa(len(roms)) + " " + "Roms"
					if e.Key == ui.KeyEnter {
						cmd := exec.Command("mame", roms[buf_init+ls.ItemPointed])
						err := cmd.Start() // fire&forget
						if err != nil {
							log.Fatal(err)
						}
					} else if e.Key == ui.KeyDelete {
						filter = ""
					} else if e.Ch == '-' && buf_init+ls.ItemPointed > 0 {
						if buf_init > 0 {
							buf_init--
							ls.Items = strs[buf_init:]
						} else {
							ls.ItemPointed--
						}
					} else if e.Ch == '+' && buf_init+ls.ItemPointed < len(roms)-1 {
						if buf_init+buf_size < len(roms) {
							buf_init++
							ls.Items = strs[buf_init:]
						} else {
							ls.ItemPointed++
						}
					} else if e.Key == ui.KeyPgup {
						if buf_init-buf_size > 0 {
							buf_init -= buf_size
						} else if buf_init > 0 {
							buf_init = 0
						} else {
							ls.ItemPointed = 0
						}
						ls.Items = strs[buf_init:]
					} else if e.Key == ui.KeyPgdn {
						if buf_init+2*buf_size < len(roms) {
							buf_init += buf_size
						} else if buf_init+buf_size < len(roms) {
							buf_init = len(strs) - buf_size
						} else {
							ls.ItemPointed = buf_size - 1
						}
						ls.Items = strs[buf_init:]
					} else if (e.Key == ui.KeyHome) && buf_init > 0 {
						ls.ItemPointed = 0
						buf_init = 0
						ls.Items = strs[buf_init:]
					} else if (e.Key == ui.KeyEnd) && buf_init+buf_size < len(strs) {
						ls.ItemPointed = buf_size - 1
						buf_init = len(strs) - buf_size
						ls.Items = strs[buf_init:]
					} else if e.Key == ui.KeyArrowUp {
						if ls.ItemPointed > 0 {
							ls.ItemPointed--
						} else if buf_init > 0 {
							buf_init--
							ls.Items = strs[buf_init:]
						}
					} else if e.Key == ui.KeyArrowDown && buf_init+ls.ItemPointed < len(strs) {
						if ls.ItemPointed < buf_size-1 {
							ls.ItemPointed++
						} else if buf_init+buf_size < len(strs) {
							buf_init++
							ls.Items = strs[buf_init:]
						}
					}
				}
			}
		default:
			ui.Body.Width = ui.TermWidth()
			ui.Body.Align()
			ls.Height = ui.TermHeight()
			ls.Width = ui.TermWidth()
			buf_size = ls.Height - 2
			ls.Items = strs[buf_init:]
			ui.Render(ls)
		}
		time.Sleep(10 * time.Millisecond)
	}

}
