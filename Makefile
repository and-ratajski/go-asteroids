.PHONY: build-local build-wasm applify run dmg serve clean

build-local:
	@go build -o dist/go-asteroids .

build-wasm:
	@sh -c 'GOOS=js GOARCH=wasm go build -o wasm/go-asteroids.wasm .'

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

serve:
	@sh -c 'cd wasm && python3 -m http.server 4000'

clean:
	@go clean
	@rm -rf dist/*
	@rm -f Asteroids.dmg
	@rm -rf "Go Asteroids.app"