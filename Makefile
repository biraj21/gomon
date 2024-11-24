bin := gomon
install_dir := /usr/local/bin

source_files := $(shell find . -name '*.go')

build: $(bin)

$(bin): $(source_files)
	go build -o $(bin)

install: $(bin)
	echo "Installing $(bin) to $(install_dir)..."
	sudo cp $(bin) $(install_dir)
	echo "Done!"

update: build install

uninstall:
	echo "Uninstalling $(bin) from $(install_dir)..."
	sudo rm $(install_dir)/$(bin)
	echo "Done!"

clean:
	rm $(bin)
