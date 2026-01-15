package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sort"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/internal/handlers"
)

const (
	cacheMaxAgeOneYear = 31536000 // 365 days in seconds
	cacheMaxAgeOneDay  = 86400    // 24 hours in seconds
)

func cacheControlMiddleware(next http.Handler, maxAge int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
		next.ServeHTTP(w, r)
	})
}

// getSortedGalleries converts gallery map to sorted slice by name
func getSortedGalleries(galleries map[string]*config.GalleryConfig) []config.GalleryMetadata {
	sortedGalleries := make([]config.GalleryMetadata, 0, len(galleries))
	for _, g := range galleries {
		sortedGalleries = append(sortedGalleries, g.Metadata)
	}

	sort.Slice(sortedGalleries, func(i, j int) bool {
		return sortedGalleries[i].Name < sortedGalleries[j].Name
	})

	return sortedGalleries
}

func main() {
	// read configuration
	serverConfig, err := config.ReadServerConfig("./configs/server.toml")
	if err != nil {
		panic(err)
	}

	// Panic on gallery config errors since galleries are core functionality
	galleries, err := config.ReadGalleryConfigsIn("./configs")
	if err != nil {
		panic(err)
	}

	// Home config is optional - gracefully degrade to text-only home page
	homeConfig, err := config.ReadHomeConfig("./configs/home.toml")
	if err != nil {
		slog.Warn("could not read home config", "error", err)
		homeConfig = &config.HomeConfig{HeroImages: []string{}}
	}

	galleriesList := getSortedGalleries(galleries)

	home := handlers.Home{
		HomeConfig: *homeConfig,
		Galleries:  galleriesList,
	}

	gallery := handlers.Galleries{
		Galleries:     galleries,
		GalleriesList: galleriesList,
	}

	about := handlers.About{
		Galleries: galleriesList,
	}

	// create routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", home.HomeHandler)
	mux.HandleFunc("GET /gallery/{shortName}", gallery.GalleryHandler)
	mux.HandleFunc("GET /about", about.AboutHandler)

	// serve static files
	static_fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", cacheControlMiddleware(static_fs, cacheMaxAgeOneYear)))

	images_fs := http.FileServer(http.Dir("./assets/images"))
	mux.Handle("GET /assets/images/", http.StripPrefix("/assets/images/", cacheControlMiddleware(images_fs, cacheMaxAgeOneDay)))

	// initialize server, if serverConfig.Port is 0 the listener will choose a random port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", serverConfig.Port))
	if err != nil {
		panic(err)
	}

	slog.Info("starting server", "port", listener.Addr().(*net.TCPAddr).Port)
	panic(http.Serve(listener, mux))
}
