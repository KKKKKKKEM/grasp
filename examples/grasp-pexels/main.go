package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/KKKKKKKEM/flowkit/core"
	"github.com/KKKKKKKEM/flowkit/x/download"
	"github.com/KKKKKKKEM/flowkit/x/extract"
	"github.com/KKKKKKKEM/flowkit/x/grasp"
	"github.com/KKKKKKKEM/flowkit/x/grasp/sites/pexels"
)

func main() {
	trackerProvider := grasp.NewMPBTrackerProvider()

	extractor := extract.NewStage("extractor")
	extractor.Mount(&pexels.APIParser{})

	downloader := download.NewStage("download")

	p := grasp.NewGraspPipeline(
		grasp.WithExtractor(extractor),
		grasp.WithDownloader(downloader),
		grasp.WithPlugin(&grasp.CLIInteractionPlugin{}),
		grasp.WithTrackerProvider(trackerProvider),
	)

	task := &grasp.Task{
		URLs: []string{"https://api.pexels.com/v1/photos/1000"},
		Extract: grasp.ExtractConfig{
			MaxRounds:   1,
			Concurrency: 1,
		},
		Download: grasp.DownloadConfig{
			Dest: ".",
		},
	}

	rc := core.NewContext(context.Background(), "pexels-example")
	rc.WithTrackerProvider(trackerProvider)
	rc.WithValue("task", task)

	runReport, err := p.Run(rc, "grasp")

	trackerProvider.Wait()

	if err != nil {
		log.Fatalf("pipeline failed: %v", err)
	}

	report := runReport.StageResults["grasp"].Outputs["report"].(*grasp.Report)
	bytes, _ := json.Marshal(report)
	log.Printf("completed in %dms, success=%v, rounds=%d, items=%d",
		report.DurationMs, report.Success, report.Rounds, report.ParsedItems)
	log.Printf("report: %s", string(bytes))
}
