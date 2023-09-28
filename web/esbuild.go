package web

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

const (
	liveReloadHost = "localhost"
	liveReloadPort = 5173 // backwards compatibility with Vite
)

func ServeAssets(ctx context.Context, baseDir string) error {
	// HACK (nachtjasmin): This is probably the dirtiest hack to move one directory up, but for now, it works.
	absDir, _ := filepath.Abs(baseDir + "/../")
	options := api.BuildOptions{
		EntryPoints:   []string{"assets/css/app.css", "assets/js/app.js"},
		Bundle:        true,
		AbsWorkingDir: absDir,
		Sourcemap:     api.SourceMapInline,
		Format:        api.FormatESModule,
		Outdir:        "assets",
		Define: map[string]string{
			"window.IS_PRODUCTION": "false",
			"window.ESBUILD_HOST":  fmt.Sprintf(`"http://%s:%d"`, liveReloadHost, liveReloadPort),
		},
	}

	buildCtx, ctxErr := api.Context(options)
	if ctxErr != nil {
		return fmt.Errorf("creating context: %w", ctxErr)
	}

	err := buildCtx.Watch(api.WatchOptions{})
	if err != nil {
		return fmt.Errorf("starting watch mode: %w", err)
	}

	_, err = buildCtx.Serve(api.ServeOptions{
		Port:     liveReloadPort,
		Host:     liveReloadHost,
		Servedir: options.Outdir,
	})
	if err != nil {
		return fmt.Errorf("serving files: %w", err)
	}

	// Cancel the build context once the parent context is done.
	go func() {
		<-ctx.Done()
		buildCtx.Cancel()
		buildCtx.Dispose()
	}()
	return nil
}
