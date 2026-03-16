package webapp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var SSL_FILES_PATH = "/etc/nginx/ssl"

func generateNginxConfig(app WebApp, certFileName string, keyFileName string) string {
	hosts := strings.Join(app.Hosts, " ")
	certPath := filepath.Join(SSL_FILES_PATH, certFileName)
	keyPath := filepath.Join(SSL_FILES_PATH, keyFileName)

	return fmt.Sprintf(`
server {
    listen %d ssl;
    server_name %s;

    ssl_certificate %s;
    ssl_certificate_key %s;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://waf_nodes;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
`, app.Port, hosts, certPath, keyPath)
}

func createNginxFiles(app WebApp, config string) {
	fmt.Println("creating nginx config for", app.ID.Hex())
	available := fmt.Sprintf("/etc/nginx/sites-available/webapp-%s.conf", app.ID.Hex())
	os.WriteFile(available, []byte(config), 0644)

	enabled := fmt.Sprintf("/etc/nginx/sites-enabled/webapp-%s.conf", app.ID.Hex())
	os.Symlink(available, enabled) // ссылка на файл

	reloadNginx()
}

func deleteNginxFiles(app WebApp) {
	available := fmt.Sprintf("/etc/nginx/sites-available/webapp-%s.conf", app.ID.Hex())
	enabled := fmt.Sprintf("/etc/nginx/sites-enabled/webapp-%s.conf", app.ID.Hex())

	os.Remove(enabled)
	os.Remove(available)
	reloadNginx()
}

func editNginxFiles(app WebApp, newConfig string) {
	fileName := fmt.Sprintf("/etc/nginx/sites-available/webapp-%s.conf", app.ID.Hex())
	os.WriteFile(fileName, []byte(newConfig), 0644)
	reloadNginx()
}

func reloadNginx() {

	cmd := exec.Command("nginx", "-t")
	if err := cmd.Run(); err != nil {
		log.Println("nginx config invalid")
		return
	}

	exec.Command("systemctl", "reload", "nginx").Run()
}
