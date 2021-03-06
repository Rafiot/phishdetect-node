// PhishDetect
// Copyright (c) 2018-2019 Claudio Guarnieri.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"time"

	// "github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type Database struct {
	Client *mongo.Client
	DB     *mongo.Database
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Key   string `json:"key"`
	Role  string `json:"role"`
}

type Indicator struct {
	Type     string    `json:"type"`
	Original string    `json:"original"`
	Hashed   string    `json:"hashed"`
	Tags     []string  `json:"tags"`
	Datetime time.Time `json:"datetime"`
	Owner    string    `json:"owner"`
}

type Event struct {
	Type        string    `json:"type"`
	Match       string    `json:"match"`
	Indicator   string    `json:"indicator"`
	UserContact string    `json:"user_contact"`
	Datetime    time.Time `json:"datetime"`
	UUID        string    `json:"uuid"`
}

func NewDatabase() (*Database, error) {
	client, err := mongo.NewClient("mongodb://localhost:27017")
	if err != nil {
		return nil, err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		return nil, err
	}
	db := client.Database("phishdetect")

	return &Database{
		Client: client,
		DB:     db,
	}, nil
}

func (d *Database) Close() {
	d.Client.Disconnect(context.Background())
}

func (d *Database) GetUsers() ([]User, error) {
	var users []User
	coll := d.DB.Collection("users")
	cur, err := coll.Find(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var user User
		if err := cur.Decode(&user); err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (d *Database) GetIndicators() ([]Indicator, error) {
	var iocs []Indicator
	coll := d.DB.Collection("indicators")
	// TODO: Need to filter output just to the bare minimum.
	cur, err := coll.Find(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var ioc Indicator
		if err := cur.Decode(&ioc); err != nil {
			continue
		}
		iocs = append(iocs, ioc)
	}

	return iocs, nil
}

func (d *Database) AddIndicator(ioc Indicator) error {
	coll := d.DB.Collection("indicators")

	var iocFind Indicator
	err := coll.FindOne(context.Background(), map[string]string{"hashed": ioc.Hashed}).Decode(&iocFind)
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
		default:
			return err
		}
	} else {
		return fmt.Errorf("This is an already known indicator")
	}

	_, err = coll.InsertOne(context.Background(), ioc)
	return err
}

func (d *Database) GetEvents() ([]Event, error) {
	coll := d.DB.Collection("events")

	// TODO: Fix sorting.
	// sortBy := []string{"-datetime", "indicator"}
	cur, err := coll.Find(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	events := []Event{}
	for cur.Next(context.Background()) {
		var event Event
		if err := cur.Decode(&event); err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func (d *Database) AddEvent(event Event) error {
	coll := d.DB.Collection("events")

	_, err := coll.InsertOne(context.Background(), event)
	return err
}
