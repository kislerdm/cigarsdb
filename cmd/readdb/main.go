package main

import (
	"cigarsdb/storage/fs"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
)

func main() {
	var name, dir string

	flag.StringVar(&name, "s", "", "cigar name to search in database")
	flag.StringVar(&dir, "p", "", "database path")
	flag.Parse()
	if name == "" || dir == "" {
		log.Println("name and dir must be provided")
		flag.Usage()
		os.Exit(1)
	}

	repository, err := fs.NewClient(dir)
	if err != nil {
		log.Fatalln(err)
	}

	r, err := repository.Seek(context.TODO(), name)
	if err != nil {
		log.Println(err)
		return
	}

	_ = json.NewEncoder(os.Stdout).Encode(r)
}
