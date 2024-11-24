bin := gomon
install_dir := /usr/local/bin

build:
	go build -o $(bin)

install: $(bin)
	echo "Installing $(bin) to $(install_dir)..."
	sudo cp $(bin) $(install_dir)
	echo "Done!"

uninstall:
	echo "Uninstalling $(bin) from $(install_dir)..."
	sudo rm $(install_dir)/$(bin)
	echo "Done!"

clean:
	rm $(bin)
