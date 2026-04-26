#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() {
  echo -e "${RED}[ERROR]${NC} $1"
  exit 1
}

if command -v pacman &>/dev/null; then
  PM="pacman"
  INSTALL="sudo pacman -S --noconfirm"
elif command -v apt &>/dev/null; then
  PM="apt"
  INSTALL="sudo apt install -y"
elif command -v dnf &>/dev/null; then
  PM="dnf"
  INSTALL="sudo dnf install -y"
else
  error "Unsupported distro. Install dependencies manually."
fi
info "Detected package manager: $PM"

SHELL_NAME=$(basename "$SHELL")
case "$SHELL_NAME" in
bash) RC="$HOME/.bashrc" ;;
zsh) RC="$HOME/.zshrc" ;;
fish) RC="$HOME/.config/fish/config.fish" ;;
*) RC="$HOME/.profile" ;;
esac
info "Detected shell: $SHELL_NAME (config: $RC)"

info "Installing make, mosquitto, avahi..."
if [ "$PM" = "pacman" ]; then
  $INSTALL make mosquitto avahi
elif [ "$PM" = "apt" ]; then
  sudo apt update
  $INSTALL make mosquitto avahi-daemon
elif [ "$PM" = "dnf" ]; then
  $INSTALL make mosquitto avahi
fi

# Go
if command -v go &>/dev/null; then
  info "Go already installed: $(go version)"
else
  info "Installing Go..."
  if [ "$PM" = "pacman" ]; then
    $INSTALL go
  elif [ "$PM" = "apt" ]; then
    $INSTALL golang-go
  elif [ "$PM" = "dnf" ]; then
    $INSTALL golang
  fi
fi

# Bun
if command -v bun &>/dev/null; then
  info "Bun already installed: $(bun --version)"
else
  info "Installing Bun..."
  curl -fsSL https://bun.sh/install | bash
  export BUN_INSTALL="$HOME/.bun"
  export PATH="$BUN_INSTALL/bin:$PATH"
fi

info "Configuring Mosquitto..."
sudo tee /etc/mosquitto/mosquitto.conf >/dev/null <<EOF
listener 1883
allow_anonymous true
EOF

info "Configuring Avahi..."
sudo tee /etc/avahi/services/mqtt.service >/dev/null <<EOF
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">
<service-group>
  <name>MQTT Broker</name>
  <service>
    <type>_mqtt._tcp</type>
    <port>1883</port>
  </service>
</service-group>
EOF

info "Starting services..."
sudo systemctl enable --now mosquitto
sudo systemctl enable --now avahi-daemon

IDF_PATH="$HOME/.espressif/tools/activate_idf_v5.5.1.sh"
ALIAS_LINE=""

if [ "$SHELL_NAME" = "fish" ]; then
  ALIAS_LINE="alias idf='source $IDF_PATH'"
else
  ALIAS_LINE="alias idf='source $IDF_PATH'"
fi

if grep -q "alias idf=" "$RC" 2>/dev/null; then
  info "idf alias already exists in $RC"
else
  echo "" >>"$RC"
  echo "# ESP-IDF" >>"$RC"
  echo "$ALIAS_LINE" >>"$RC"
  info "Added idf alias to $RC"
fi

echo ""
info "Checking services..."
sudo systemctl is-active mosquitto && info "Mosquitto:    OK" || warning "Mosquitto:    FAILED"
sudo systemctl is-active avahi-daemon && info "Avahi:        OK" || warning "Avahi:        FAILED"
command -v go &>/dev/null && info "Go:           OK ($(go version))" || warning "Go:           NOT FOUND"
command -v bun &>/dev/null && info "Bun:          OK ($(bun --version))" || warning "Bun:          NOT FOUND"
command -v make &>/dev/null && info "Make:         OK" || warning "Make:         NOT FOUND"

info "Installing necessary packages..."

# Go dependencies
go mod tidy
info "Go modules: OK"

# Frontend dependencies
cd web && bun i && cd ..
info "Bun install: OK"

info "Done! Installed all necessary packages!"

echo ""
warning "ESP-IDF must be installed manually via EIM (GUI installer):"
echo "  https://docs.espressif.com/projects/idf-im-ui/en/latest/"
echo ""
warning "After installing, run the following to apply the idf alias:"
echo ""
echo "  source $RC"
echo ""
info "Then use 'idf' to load ESP-IDF in any terminal session."
