🌐 URL Downloader CLI

A fast, reliable, and portable command-line tool for downloading files from the internet.
Built with Go, it provides a clean interface, real-time progress feedback, and seamless Docker support.

✨ Features

📥 Direct Downloads – Fetch files from any valid URL.

📊 Progress Bar – Monitor real-time download progress.

💾 Flexible Output – Save files to the current directory or a custom path.

🐳 Dockerized – Run in isolated environments without installing Go.

⚡ Lightweight & Fast – Small binary, minimal dependencies.

📦 Installation
# Clone the repository
git clone https://github.com/franzego/url-downloader.git
cd url-downloader

# Build the binary
go build -o dazai .

🔹 Using Docker


⚙️ Command-Line Options
| Flag | Description                                         | Example                           |
| ---- | --------------------------------------------------- | --------------------------------- |
| `-u` | **(Required)** URL of the file to download          | `-u https://example.com/file.zip` |
| `-o` | Output directory or filepath (default: current dir) | `-o ./downloads/`                 |
| `-h` | Show help and usage information                     | `./urldownloader -h`              |

📂 Example Workflow

# Download an image
./urldownloader -u https://example.com/picture.png

# Save an archive to a custom folder
./urldownloader -u https://example.com/archive.tar.gz -o ~/Downloads/
✅ You’ll see a live progress bar while downloading.

🐳 Dockerfile Example


🤝 Contributing

Contributions are welcome! 🚀

Open an issue for bugs or feature requests.

Submit a pull request with improvements.

Please follow conventional commit messages and ensure all code is properly tested.

📜 License

Released under the MIT License.
You’re free to use, modify, and distribute this tool.

