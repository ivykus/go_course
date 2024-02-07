package main

import (
	"context"
	"log"
	"time"

	pb "github.com/ivykus/gocourse/mailinglist/proto"
)

func logResponse(res *pb.EmailResponse, err error) {
	if err != nil {
		log.Fatalf("  error: %v", err)
	}
	if res == nil {
		log.Print("  email not found")
	} else {
		log.Printf("  response: %v", res.EmailEntry)
	}
}

func createEmail(client pb.MailingListServiceClient, addr string) {
	log.Println("create email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.CreateEmail(ctx, &pb.CreateEmailRequest{EmailAddr: addr})
	logResponse(res, err)
}

func getEmail(client pb.MailingListServiceClient, addr string) {
	log.Println("get email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmail(ctx, &pb.GetEmailRequest{EmailAddr: addr})
	logResponse(res, err)
}
func GetEmailBatch(client pb.MailingListServiceClient, count int, page int) {
	log.Println("get email batch")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmailBatch(ctx, &pb.GetEmailBatchRequest{
		Page:  int32(page),
		Count: int32(count),
	})
	if err != nil {
		log.Fatalf("  error: %v", err)
	}
	log.Printf("  response:")
	for i := 0; i < len(res.EmailEntries); i++ {
		log.Printf("    item [%v of %v]: %v",
			i+1, len(res.EmailEntries), res.EmailEntries[i])
	}

}
