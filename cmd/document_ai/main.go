package main

import (
	"context"
	"fmt"

	documentai "cloud.google.com/go/documentai/apiv1"
	documentaipb "cloud.google.com/go/documentai/apiv1/documentaipb"
)

func main() {
	ctx := context.Background()

	c, err := documentai.NewDocumentProcessorClient(ctx)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	req := &documentaipb.BatchProcessRequest{
		Name: "projects/1000702815736/locations/us/processors/a08c28658c0f5f87/processorVersions/pretrained-expense-v1.4-2022-11-18",
		InputDocuments: &documentaipb.BatchDocumentsInputConfig{
			Source: &documentaipb.BatchDocumentsInputConfig_GcsDocuments{
				GcsDocuments: &documentaipb.GcsDocuments{
					Documents: []*documentaipb.GcsDocument{
						{
							GcsUri:   "gs://expense-receipts/2023_10_11__giant__53_00.pdf",
							MimeType: "application/pdf",
						},
					},
				},
			},
		},
		DocumentOutputConfig: &documentaipb.DocumentOutputConfig{
			Destination: &documentaipb.DocumentOutputConfig_GcsOutputConfig_{
				GcsOutputConfig: &documentaipb.DocumentOutputConfig_GcsOutputConfig{
					GcsUri: "gs://expense-receipts/output",
				},
			},
		},
		SkipHumanReview: false,
	}

	op, err := c.BatchProcessDocuments(ctx, req)
	if err != nil {
		panic(err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		panic(err)
	}
	// TODO: Use resp.
	fmt.Println(resp)
}
