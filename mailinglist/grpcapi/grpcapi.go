package grpcapi

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	"github.com/ivykus/gocourse/mailinglist/mdb"
	pb "github.com/ivykus/gocourse/mailinglist/proto"
	"google.golang.org/grpc"
)

type MailServer struct {
	pb.UnimplementedMailingListServiceServer
	db *sql.DB
}

func pbEntryToMdbEntry(pbEntry *pb.EmailEntry) mdb.EmailEntry {
	t := time.Unix(pbEntry.ConfirmedAt, 0)
	return mdb.EmailEntry{
		Id:          pbEntry.Id,
		Email:       pbEntry.Email,
		ConfirmedAt: &t,
		OptOut:      pbEntry.OptOut,
	}
}

func mdbEntryToPbEntry(mdbEntry *mdb.EmailEntry) pb.EmailEntry {
	return pb.EmailEntry{
		Id:          mdbEntry.Id,
		Email:       mdbEntry.Email,
		ConfirmedAt: mdbEntry.ConfirmedAt.Unix(),
		OptOut:      mdbEntry.OptOut,
	}
}

func emailResponse(db *sql.DB, email string) (*pb.EmailResponse, error) {
	entry, err := mdb.GetEmail(db, email)
	if err != nil {
		return &pb.EmailResponse{}, err
	}
	if entry == nil {
		return &pb.EmailResponse{}, nil
	}

	res := mdbEntryToPbEntry(entry)
	return &pb.EmailResponse{EmailEntry: &res}, nil
}

func (s *MailServer) GetEmail(ctx context.Context, r *pb.GetEmailRequest) (*pb.EmailResponse, error) {
	log.Println("gRPC GetEmail: ", r)
	return emailResponse(s.db, r.EmailAddr)
}

func (s *MailServer) GetEmailBatch(
	ctx context.Context, req *pb.GetEmailBatchRequest) (*pb.GetEmailBatchResponse, error) {
	log.Println("gRPC GetEmailBatch: ", req)
	mdbEntries, err := mdb.GetEmailBatch(s.db, mdb.GetEmailBatchQueryParams{
		Count: int(req.Count),
		Page:  int(req.Page),
	})
	if err != nil {
		return &pb.GetEmailBatchResponse{}, err
	}
	pbEntries := make([]*pb.EmailEntry, 0, len(mdbEntries))
	for _, email := range mdbEntries {
		entry := mdbEntryToPbEntry(&email)
		pbEntries = append(pbEntries, &entry)
	}
	return &pb.GetEmailBatchResponse{EmailEntries: pbEntries}, nil
}

func (s *MailServer) CreateEmail(ctx context.Context, r *pb.CreateEmailRequest) (*pb.EmailResponse, error) {
	log.Println("gRPC CreateEmail: ", r)

	err := mdb.CreateEmail(s.db, r.EmailAddr)
	if err != nil {
		return &pb.EmailResponse{}, err
	}

	return emailResponse(s.db, r.EmailAddr)
}

func (s *MailServer) UpdateEmail(ctx context.Context, r *pb.UpdateEmailRequest) (*pb.EmailResponse, error) {
	log.Println("gRPC UpdateEmail: ", r)

	entry := pbEntryToMdbEntry(r.EmailEntry)
	err := mdb.UpdateEmail(s.db, &entry)
	if err != nil {
		return &pb.EmailResponse{}, err
	}

	return emailResponse(s.db, entry.Email)
}

func (s *MailServer) DeleteEmail(ctx context.Context, r *pb.DeleteEmailRequest) (*pb.EmailResponse, error) {
	log.Println("gRPC DeleteEmail: ", r)

	err := mdb.DeleteEmail(s.db, r.EmailAddr)
	if err != nil {
		return &pb.EmailResponse{}, err
	}

	return emailResponse(s.db, r.EmailAddr)
}

func Serve(db *sql.DB, bind string) {
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("gRPC server error: %v\n", err)
	}

	grpcServer := grpc.NewServer()
	mailServer := MailServer{db: db}
	pb.RegisterMailingListServiceServer(grpcServer, &mailServer)

	log.Printf("gRPC API server listening on %s\n", bind)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("gRPC server error: %v\n", err)
	}
}
