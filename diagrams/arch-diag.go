package main

import (
	"log"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/apps"
	"github.com/blushft/go-diagrams/nodes/generic"
)

func main() {
	d, err := diagram.New(diagram.Filename("diagram"), diagram.Label("App"), diagram.Direction("LR"))
	if err != nil {
		log.Fatal(err)
	}

	user := apps.Client.User(diagram.NodeLabel("User"))
	client := generic.Device.Mobile(diagram.NodeLabel("i9lyfe frontend"))
	db := apps.Database.Postgresql(diagram.NodeLabel("Database"))
	cacheDB := apps.Inmemory.Redis(diagram.NodeLabel("Cache\nIn-memory DB"))
	queue := apps.Inmemory.Redis(diagram.NodeLabel("Event Queue\nBg Task Queue\n(Redis stream)"))

	apihand := diagram.NewNode(diagram.NodeLabel("API Controllers\n(HTTP request handlers)"), diagram.LabelLocation("c"), diagram.FixedSize(false))
	wsconnhand := diagram.NewNode(diagram.NodeLabel("Realtime Controller\n(WS connections handler)"), diagram.LabelLocation("c"), diagram.FixedSize(false))
	wsmsghand := diagram.NewNode(diagram.NodeLabel("API Controllers\n(WS message handlers)"), diagram.LabelLocation("c"), diagram.FixedSize(false))
	apiservices := diagram.NewNode(diagram.NodeLabel("API services"), diagram.LabelLocation("c"), diagram.FixedSize(false))
	models := diagram.NewNode(diagram.NodeLabel("Models"), diagram.LabelLocation("c"), diagram.FixedSize(false))
	bgprocesses := diagram.NewNode(diagram.NodeLabel("Background processes"), diagram.LabelLocation("c"), diagram.FixedSize(false))

	realtimeservice := diagram.NewNode(diagram.NodeLabel("Realtime service"), diagram.LabelLocation("c"), diagram.FixedSize(false))

	backapi := diagram.NewGroup("backapi").Label("Backend API").
		Add(apihand).
		Add(wsconnhand).
		Add(wsmsghand).
		Add(apiservices).
		Add(models).
		Add(realtimeservice).
		Add(bgprocesses).
		Connect(wsconnhand, wsmsghand, diagram.Bidirectional()).
		Connect(apihand, apiservices, diagram.Bidirectional()).
		Connect(wsmsghand, apiservices, diagram.Bidirectional()).
		Connect(apiservices, models, diagram.Bidirectional()).
		Connect(apiservices, realtimeservice).
		Connect(wsconnhand, realtimeservice)

	d.Connect(user, client, diagram.Forward()).
		Connect(client, apihand, diagram.Bidirectional()).
		Connect(client, wsconnhand, diagram.Bidirectional()).
		Connect(apiservices, queue).
		Connect(models, db, diagram.Bidirectional()).
		Connect(models, cacheDB, diagram.Bidirectional()).
		Connect(bgprocesses, queue, diagram.Reverse()).
		Connect(bgprocesses, cacheDB).
		Group(backapi)

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}
