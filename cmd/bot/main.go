package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anoop142/zteScanner"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	_ "modernc.org/sqlite"
)

type config struct {
	username string
	password string
	url      string
}

func printUsage() {
	fmt.Printf("%s -db <sqlite db path> -u <router username> -p <router password>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	cfg := config{}
	dbPath := flag.String("db", "", "sqlite db file")
	flag.StringVar(&cfg.username, "u", "", "router username")
	flag.StringVar(&cfg.password, "p", "", "router password")
	flag.StringVar(&cfg.url, "url", "http://192.168.1.1", "router URL")

	flag.Usage = printUsage
	flag.Parse()

	if *dbPath == "" || cfg.username == "" || cfg.password == "" {
		printUsage()
		os.Exit(1)
	}
	botToken := os.Getenv("ZTE_BOT_TOKEN")
	// "," separated admin ids
	adminID := os.Getenv("ZTE_BOT_ADMIN_ID")

	if botToken == "" {
		log.Fatal("empty bot token: ZTE_BOT_TOKEN")
	}

	if adminID == "" {
		log.Fatal("empty admin ID: ZTE_BOT_ADMIN_ID")
	}

	var adminIDs []int64
	for _, v := range strings.Split(adminID, ",") {
		a, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		adminIDs = append(adminIDs, a)
	}

	if len(adminIDs) == 0 {
		log.Fatal("no telegram adminID supplied")
	}

	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	if _, err := os.Stat(*dbPath); err != nil {
		// init db if file not found
		if os.IsNotExist(err) {
			err = initDB(*dbPath)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}

	db, err := openDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}

	models := NewModels(db)

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	commands := []tele.Command{
		tele.Command{
			Text:        "get_devs",
			Description: "get list of all devices",
		},
		tele.Command{
			Text:        "get_devs_alive",
			Description: "get list of alive devices",
		},
		tele.Command{
			Text:        "help",
			Description: "show help",
		},
	}

	b.SetCommands(commands)

	adminOnly := b.Group()
	adminOnly.Use(middleware.Whitelist(adminIDs...))

	scanner := zteScanner.New(cfg.url, cfg.username, cfg.password)

	b.Handle(tele.OnText, func(c tele.Context) error {
		var (
			user = c.Sender()
			text = c.Text()
		)

		log.Println(user.FirstName, ":", user.ID, ":", text)

		return nil
	})

	adminOnly.Handle("/get_devs", func(c tele.Context) error {
		devs, err := getDevsUsingDB(scanner, models)
		if err != nil {
			return c.Send("ERROR: " + err.Error())
		}
		return c.Send(devs)

	})

	adminOnly.Handle("/get_devs_alive", func(c tele.Context) error {

		devs, err := getDevsAliveUsingDB(scanner, models)
		if err != nil {
			return c.Send("ERROR: " + err.Error())
		}
		return c.Send(devs)

	})

	adminOnly.Handle("/get_devs_all", func(c tele.Context) error {

		devs, err := getDevsAll(scanner)
		if err != nil {
			return c.Send("ERROR: " + err.Error())
		}
		return c.Send(devs)

	})

	adminOnly.Handle("/save", func(c tele.Context) error {
		args := c.Args()
		if len(args) != 2 {
			return c.Send("not enough args: /save <mac> <name>\n")
		}
		mac := args[0]
		name := args[1]
		err := saveDevice(models, mac, name)
		if err != nil {
			return c.Send("error: " + err.Error() + "\n")
		}
		return c.Send(mac + " saved!\n")
	})

	adminOnly.Handle("/ignore", func(c tele.Context) error {
		args := c.Args()
		if len(args) != 2 {
			return c.Send("not enough args: /ignore <mac> <name>\n")
		}
		mac := args[0]
		name := args[1]
		err := ignoreDevice(models, mac, name)
		if err != nil {
			return c.Send("error: " + err.Error() + "\n")
		}
		return c.Send(mac + " added to ignore list!\n")
	})

	adminOnly.Handle("/list_saved", func(c tele.Context) error {
		devs, err := getSavedList(models)
		if err != nil {
			return c.Send("error: " + err.Error() + "\n")
		}
		return c.Send(devs)
	})

	adminOnly.Handle("/list_ignored", func(c tele.Context) error {
		devs, err := getIgnoredList(models)
		if err != nil {
			return c.Send("error: " + err.Error() + "\n")
		}
		return c.Send(devs)
	})

	adminOnly.Handle("/delete", func(c tele.Context) error {
		args := c.Args()
		if len(args) != 1 || args[0] == "" {
			return c.Send("not enough args: /delete <mac>\n")
		}

		mac := args[0]
		err := models.KnownDevices.Delete(mac)

		if err != nil {
			return c.Send("error: " + err.Error() + "\n")
		}
		return c.Send(mac + " deleted!")
	})

	adminOnly.Handle("/help", func(c tele.Context) error {
		helpMsg := "/get_devs:  list all devices \n"
		helpMsg += "/get_devs_alive: list ping-able devices\n"
		helpMsg += "/get_devs_all: list all devices without using DB\n"
		helpMsg += "/list_saved: list saved devices\n"
		helpMsg += "/list_ignored: list ignored devices\n"
		helpMsg += "/save <mac> <name>: map a device to name\n"
		helpMsg += "/ignore <mac> <name>: ignore the device \n"
		helpMsg += "/delete <mac>: delete saved/ignored device from db\n"
		return c.Send(helpMsg)
	})

	b.Start()
}
