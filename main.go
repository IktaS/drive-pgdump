package main

import (
	"context"
	"log"
	"os"
	"strconv"

	pg "github.com/habx/pg-commands"
	"github.com/joho/godotenv"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

func main() {
	srv, err := drive.NewService(context.Background(), option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalf("unable to access Drive API: %v", err)
	}
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("failed reading .env file: %v", err)
	}
	dump, err := pg.NewDump(&pg.Postgres{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		DB:       os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
	})
	if err != nil {
		log.Fatalf("failed conncting to DB: %v", err)
	}
	dumpExec := dump.Exec(pg.ExecOptions{StreamPrint: false})
	if dumpExec.Error != nil {
		os.Remove(dumpExec.File)
		log.Println(dumpExec.Output)
		log.Fatalf("failed dumping: %v", dumpExec.Error.Err)
	}
	log.Println("Dump success")
	log.Println(dumpExec.Output)

	filename := dumpExec.File

	goFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening %q: %v", filename, err)
	}
	_, err = srv.Files.Insert(&drive.File{Title: filename, Parents: []*drive.ParentReference{
		{
			Id:     os.Getenv("DRIVE_FOLDER_ID"),
			IsRoot: true,
		},
	}}).Media(goFile).Do()
	if err != nil {
		log.Fatalf("error inserting %q: %v", filename, err)
	}
	log.Printf("Uploaded %s", filename)
	os.Remove(filename)
}
