package main

import (
	"context"
	"log"
	"time"

	"github.com/alexflint/go-arg"
	pb "github.com/ivykus/gocourse/mailinglist/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func createEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("create email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.CreateEmail(ctx, &pb.CreateEmailRequest{EmailAddr: addr})
	logResponse(res, err)
	return res.EmailEntry
}

func getEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("get email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmail(ctx, &pb.GetEmailRequest{EmailAddr: addr})
	logResponse(res, err)
	return res.EmailEntry
}
func updateEmail(client pb.MailingListServiceClient, entry *pb.EmailEntry) *pb.EmailEntry {
	log.Println("update email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.UpdateEmail(ctx, &pb.UpdateEmailRequest{EmailEntry: entry})
	logResponse(res, err)
	return res.EmailEntry
}

func deleteEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("delete email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.DeleteEmail(ctx, &pb.DeleteEmailRequest{EmailAddr: addr})
	logResponse(res, err)
	return res.EmailEntry
}
func getEmailBatch(client pb.MailingListServiceClient, count int, page int) {
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

var args struct {
	GrpcAddr string `arg:"env:MAILINGLIST_GRPC_ADDR"`
}

func main() {
	arg.MustParse(&args)
	if args.GrpcAddr == "" {
		args.GrpcAddr = ":8081"
	}

	conn, err := grpc.Dial(args.GrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewMailingListServiceClient(conn)

	newEmail := createEmail(client, "bob@alice.got")
	newEmail.ConfirmedAt = 666
	updateEmail(client, newEmail)
	deleteEmail(client, "bob@alice.got")

	getEmailBatch(client, 50, 1)
}
