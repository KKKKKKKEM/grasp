package main

import (
	"log"

	"github.com/KKKKKKKEM/flowkit/x/grasp"
)

func main() {
	p := grasp.NewGraspPipeline()

	err := p.CLI()
	if err != nil {
		log.Fatal(err)
	}

	if err := p.Serve(":8080"); err != nil {
		log.Fatal(err)
	}

}
