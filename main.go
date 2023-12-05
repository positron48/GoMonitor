package main

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	_ "net/smtp"
)

type Pattern struct {
	Regexp string
	Email  string
}

var patterns []Pattern

func main() {
	// Init patterns by something
	patterns = append(patterns, Pattern{Regexp: ".*", Email: "positron48@gmail.com"})

	// Connect to Elasticsearch
	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
	if err != nil {
		panic(err)
	}

	// for each pattern query logs and send email
	for _, pattern := range patterns {
		logs, err := fetchLogs(client, pattern.Regexp)
		if err != nil {
			panic(err)
		}

		for _, log := range logs {
			err := sendEmail(pattern.Email, log)
			if err != nil {
				panic(err)
			}
		}
	}
}

func fetchLogs(client *elastic.Client, regexp string) ([]string, error) {

	// Search with a term query
	termQuery := elastic.NewRegexpQuery("message", regexp)
	searchResult, err := client.Search().
		Index("filebeat-*").     // search in index filebeat-*
		Query(termQuery).        // specify the query
		From(0).Size(10).        // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		return nil, err
	}

	// Iterate through results
	var logs []string
	for _, hit := range searchResult.Hits.Hits {
		var log map[string]interface{}
		err := json.Unmarshal(hit.Source, &log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log["message"].(string))
	}
	return logs, nil
}

func sendEmail(to string, log string) error {
	// Set up authentication information.
	//auth := smtp.PlainAuth(
	//	"",
	//	"",
	//	"",
	//	"smtp.gmail.com",
	//)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	msg := []byte("To: " + to + "\r\n" +
		"Subject: Log Alert!" + "\r\n" +
		"\r\n" +
		log + "\r\n")

	// just print msg to console
	println(msg)
	return nil
	//return smtp.SendMail("smtp.gmail.com:587", auth, "positron48@gmail.com", []string{to}, msg)
}
