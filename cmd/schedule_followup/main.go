package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"example.com/order-follow-up-delay"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("usage: schedule_followup ORDER_ID CUSTOMER_ID HOURS")
	}
	hours, err := strconv.Atoi(os.Args[3])
	if err != nil || hours < 1 {
		log.Fatal("HOURS must be a positive integer")
	}
	client, err := followup.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	followUp := followup.Schedule(os.Args[1], os.Args[2], hours, time.Now())
	messageID, err := client.Publish(context.Background(), followUp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("queued %s for %s\n", messageID, followUp.DueAt.Format(time.RFC3339))
}
