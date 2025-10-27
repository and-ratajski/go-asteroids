.PHONY: build-local build-wasm applify run dmg serve-wasm serve-pages clean deploy-pages

build-local:
	@go build -o dist/go-asteroids .
	@echo "✅ Local build complete! Files ready in dist/"

build-wasm:
	@GOOS=js GOARCH=wasm go build -ldflags="-s -w" -trimpath -o wasm/go-asteroids.wasm .

# Need to install it first, see https://github.com/machinebox/appify
applify:
	@applify -author "And Ratajski" -version "v.1.0.0" -name "Go Asteroids" -icon "./assets/images/icon.png" ./dist/go-asteroids

run:
	@go run .

dmg:
	@make applify
	@rm -rf dist/Go\ Asteroids.app
	@mv "Go Asteroids.app" dist
	@rm -f dist/go-asteroids
	@hdiutil create -volname "Go Asteroids" -srcfolder dist -ov -format UDZO Asteroids.dmg

serve-wasm:
	@make build-wasm
	@sh -c 'cd wasm && python3 -m http.server 4000'

serve-pages:
	@make build-wasm
	@cp wasm/go-asteroids.wasm docs/
	@sh -c 'cd docs && python3 -m http.server 4000'

deploy-pages:
	@GOOS=js GOARCH=wasm go build -ldflags="-s -w" -trimpath -o docs/go-asteroids.wasm .

clean:
	@go clean
	@rm -rf dist/*
	@rm -f Asteroids.dmg
	@rm -rf "Go Asteroids.app"