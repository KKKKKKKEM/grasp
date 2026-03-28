package main

import (
	"log"

	"github.com/KKKKKKKEM/flowkit/x/download"
	"github.com/KKKKKKKEM/flowkit/x/extract"
	"github.com/KKKKKKKEM/flowkit/x/grasp"
	"github.com/KKKKKKKEM/flowkit/x/grasp/sites/pexels"
)

func main() {
	extractor := extract.NewStage("extractor")
	extractor.Mount(&pexels.APIParser{})

	downloader := download.NewStage("download")

	p := grasp.NewGraspPipeline(
		grasp.WithExtractor(extractor),
		grasp.WithDownloader(downloader),
	)

	log.Println("listening on :8080")
	if err := p.Serve(":8080"); err != nil {
		log.Fatal(err)
	}
}
