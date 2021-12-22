package week03

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"
)

func startServer(srv *http.Server) error {
	http.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("test")
	})
	err := srv.ListenAndServe()
	return err
}

func ErrGroupJob() {

	group, errCtx := errgroup.WithContext(context.Background())

	srv := &http.Server{Addr: ":9000"}

	group.Go(func() error {
		return startServer(srv)
	})

	group.Go(func() error {
		<-errCtx.Done()
		return srv.Shutdown(errCtx)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	group.Go(func() error {
		for {
			select {
			case <-errCtx.Done():
				return errCtx.Err()
			case <-sigChan:
				return errors.New("get os signal: %v")
			}

		}
	})

	if err := group.Wait(); err != nil {
		fmt.Println("errgroup down")
	}

}
