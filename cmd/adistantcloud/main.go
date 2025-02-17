package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/internal/handlers"
)

func main() {
	// read configuration
	serverConfig, err := config.ReadServerConfig("./configs/server.toml")
	if err != nil {
		panic(err)
	}

	galleryConfig, err := config.ReadGalleryConfig("./configs/gallery.toml")
	if err != nil {
		panic(err)
	}

	gallery := handlers.Gallery{
		GalleryConfig:  *galleryConfig,
		NumRowsPerPage: serverConfig.NumRowsPerGalleryPage,
	}

	// create routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", gallery.GalleryHandler)
	mux.HandleFunc("GET /img/{index}", gallery.ImagePageHandler)
	mux.HandleFunc("GET /about", handlers.AboutHandler)

	// serve static files
	static_fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", static_fs))

	images_fs := http.FileServer(http.Dir("./assets/images"))
	mux.Handle("GET /assets/images/", http.StripPrefix("/assets/images/", images_fs))

	// initialize server, if serverConfig.Port is 0 the listener will choose a random port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", serverConfig.Port))
	if err != nil {
		panic(err)
	}

	slog.Info("starting server", "port", listener.Addr().(*net.TCPAddr).Port)
	panic(http.Serve(listener, mux))
}
