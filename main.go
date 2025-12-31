package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "login":
		handleLogin()
	case "connect":
		handleConnect()
	case "logout":
		handleLogout()
	case "status":
		handleStatus()
	case "info":
		handleInfo()
	case "help":
		printUsage()
	default:
		fmt.Printf("Comando desconocido: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// executeCommand ejecuta un comando del sistema
func executeCommand(cmdString string) error {
	parts := strings.Fields(cmdString)
	if len(parts) == 0 {
		return fmt.Errorf("comando vacío")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printUsage() {
	fmt.Println("GoNauta - Cliente CLI para Nauta")
	fmt.Println("\nUso: gonauta <comando> [opciones]")
	fmt.Println("\nComandos disponibles:")
	fmt.Println("  login [--vpn] - Guardar credenciales (usuario y contraseña)")
	fmt.Println("                  --vpn: Configurar comandos de VPN")
	fmt.Println("  connect       - Iniciar sesión en Nauta (ejecuta VPN automáticamente si está configurado)")
	fmt.Println("  logout        - Cerrar sesión activa (desconecta VPN automáticamente si está configurado)")
	fmt.Println("  status        - Ver tiempo restante de la sesión activa")
	fmt.Println("  info          - Ver información completa del usuario")
	fmt.Println("  help          - Mostrar esta ayuda")
	fmt.Println("\nConfiguración de VPN:")
	fmt.Println("  Los comandos VPN se ejecutan automáticamente:")
	fmt.Println("  - Conexión VPN: Después de conectar a Nauta")
	fmt.Println("  - Desconexión VPN: Antes de cerrar sesión (con delay de 2 segundos)")
	fmt.Println("\nEjemplos de comandos VPN (NordVPN):")
	fmt.Println("  Windows:")
	fmt.Println("    Conexión:    C:\\Program Files\\NordVPN\\nordvpn.exe -c")
	fmt.Println("    Desconexión: C:\\Program Files\\NordVPN\\nordvpn.exe -d")
	fmt.Println("  Linux/macOS:")
	fmt.Println("    Conexión:    nordvpn connect")
	fmt.Println("    Desconexión: nordvpn disconnect")
}

func handleLogin() {
	reader := bufio.NewReader(os.Stdin)

	// Verificar si se pasó el flag --vpn
	configureVPN := false
	if len(os.Args) > 2 && os.Args[2] == "--vpn" {
		configureVPN = true
	}

	fmt.Print("Usuario (ej: usuario@nauta.com.cu): ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Println("Error: El usuario no puede estar vacío")
		os.Exit(1)
	}

	fmt.Print("Contraseña: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Printf("Error leyendo contraseña: %v\n", err)
		os.Exit(1)
	}

	password := string(passwordBytes)
	if password == "" {
		fmt.Println("Error: La contraseña no puede estar vacía")
		os.Exit(1)
	}

	var vpnConnectCmd, vpnDisconnectCmd string

	if configureVPN {
		fmt.Println("\n--- Configuración de VPN ---")
		fmt.Print("Comando de conexión VPN: ")
		vpnConnectCmd, _ = reader.ReadString('\n')
		vpnConnectCmd = strings.TrimSpace(vpnConnectCmd)

		fmt.Print("Comando de desconexión VPN: ")
		vpnDisconnectCmd, _ = reader.ReadString('\n')
		vpnDisconnectCmd = strings.TrimSpace(vpnDisconnectCmd)
	}

	if err := SaveCredentials(username, password, vpnConnectCmd, vpnDisconnectCmd); err != nil {
		fmt.Printf("Error guardando credenciales: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Credenciales guardadas exitosamente")
	if configureVPN && (vpnConnectCmd != "" || vpnDisconnectCmd != "") {
		fmt.Println("✓ Configuración de VPN guardada")
	}
	fmt.Println("  Use 'gonauta connect' para iniciar sesión")
}

func handleConnect() {
	// Verificar si ya existe una sesión activa
	existingSession, err := LoadSession()
	if err == nil && existingSession != nil {
		fmt.Println("⚠️  Ya existe una sesión activa")
		fmt.Printf("  Usuario: %s\n", existingSession.Username)
		fmt.Println("\nUse 'gonauta status' para ver el tiempo restante")
		fmt.Println("Use 'gonauta logout' para cerrar la sesión actual antes de conectar nuevamente")
		os.Exit(0)
	}

	config, err := LoadCredentials()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Use 'gonauta login' para guardar sus credenciales primero")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fmt.Printf("Error creando cliente: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Conectando a Nauta...")
	session, err := client.Login(config.Username, config.Password)
	if err != nil {
		fmt.Printf("Error al iniciar sesión: %v\n", err)
		os.Exit(1)
	}

	if err := SaveSession(session); err != nil {
		fmt.Printf("Advertencia: No se pudo guardar la sesión: %v\n", err)
	}

	fmt.Println("✓ Sesión iniciada exitosamente")
	fmt.Printf("  Usuario: %s\n", session.Username)

	// Ejecutar comando de conexión VPN si está configurado
	if config.VPNConnectCmd != "" {
		fmt.Println("\nConectando VPN...")
		if err := executeCommand(config.VPNConnectCmd); err != nil {
			fmt.Printf("⚠️  Error ejecutando comando VPN: %v\n", err)
		} else {
			fmt.Println("✓ VPN conectado")
		}
	}

	fmt.Println("\nUse 'gonauta status' para ver el tiempo restante")
	fmt.Println("Use 'gonauta logout' para cerrar la sesión")
}

func handleLogout() {
	sessionData, err := LoadSession()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Cargar configuración para obtener comandos VPN
	config, err := LoadCredentials()
	if err != nil {
		fmt.Printf("Error cargando configuración: %v\n", err)
		os.Exit(1)
	}

	// Verificar si está conectado a través de VPN
	fmt.Println("Verificando conexión...")
	ipInfo, err := checkConnection()
	if err != nil {
		fmt.Printf("Error verificando conexión: %v\n", err)
		os.Exit(1)
	}

	// Si está conectado desde fuera de Cuba (VPN detectado)
	if ipInfo.CountryCode != "CU" {
		// Verificar si hay comando de desconexión VPN configurado
		if config.VPNDisconnectCmd != "" {
			// Ejecutar desconexión automática
			fmt.Printf("\n⚠️  Conectado a través de VPN (País: %s, ISP: %s)\n", ipInfo.Country, ipInfo.ISP)
			fmt.Println("Desconectando VPN...")
			time.Sleep(2 * time.Second) // Delay de 2 segundos
			if err := executeCommand(config.VPNDisconnectCmd); err != nil {
				fmt.Printf("⚠️  Error ejecutando comando VPN: %v\n", err)
			} else {
				fmt.Println("✓ VPN desconectado")
			}
		} else {
			// No hay comando de desconexión configurado
			fmt.Printf("\n⚠️  Conectado a través de VPN\n")
			fmt.Printf("País: %s\n", ipInfo.Country)
			fmt.Printf("ISP: %s\n\n", ipInfo.ISP)
			fmt.Println("Para desconexión automática de VPN, configure un comando de desconexión usando:")
			fmt.Println("  gonauta login --vpn")
			fmt.Println("\nNo se puede cerrar sesión mientras esté conectado a través de VPN sin comando de desconexión configurado.")
			os.Exit(1)
		}
	}

	client, err := NewClient()
	if err != nil {
		fmt.Printf("Error creando cliente: %v\n", err)
		os.Exit(1)
	}

	session := NewSession(*sessionData, client)

	fmt.Println("Cerrando sesión...")
	err = session.Logout()
	if err != nil {
		fmt.Printf("Error al cerrar sesión: %v\n", err)
		os.Exit(1)
	}

	if err := DeleteSession(); err != nil {
		fmt.Printf("Advertencia: No se pudo eliminar el archivo de sesión: %v\n", err)
	}

	fmt.Println("✓ Sesión cerrada exitosamente")
}

func handleStatus() {
	sessionData, err := LoadSession()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Use 'gonauta connect' para iniciar sesión primero")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fmt.Printf("Error creando cliente: %v\n", err)
		os.Exit(1)
	}

	session := NewSession(*sessionData, client)

	remainingTime, err := session.GetRemainingTime()
	if err != nil {
		fmt.Printf("Error obteniendo tiempo restante: %v\n", err)
		// Solo mostrar mensaje de sesión expirada si no es un error de VPN
		if !strings.Contains(err.Error(), "VPN") {
			fmt.Println("\nLa sesión puede haber expirado. Use 'gonauta connect' para iniciar sesión nuevamente")
		}
		os.Exit(1)
	}

	fmt.Printf("⏱  Tiempo restante: %02d:%02d:%02d\n",
		remainingTime.Hours,
		remainingTime.Minutes,
		remainingTime.Seconds)
}

func handleInfo() {
	config, err := LoadCredentials()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Use 'gonauta login' para guardar sus credenciales primero")
		os.Exit(1)
	}

	client, err := NewClient()
	if err != nil {
		fmt.Printf("Error creando cliente: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Obteniendo información del usuario...")
	userInfo, err := client.GetUserInfo(config.Username, config.Password)
	if err != nil {
		fmt.Printf("Error obteniendo información: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Información del Usuario ===")
	fmt.Printf("Estado: %s\n", userInfo.Status)
	fmt.Printf("Créditos: %.2f CUP\n", userInfo.Credits)
	fmt.Printf("Fecha de expiración: %s\n", userInfo.ExpirationDate)
	fmt.Printf("Tipo de acceso: %s\n", userInfo.AccessInfo)
	fmt.Printf("Tiempo disponible: %02d:%02d:%02d\n",
		userInfo.RemainingTime.Hours,
		userInfo.RemainingTime.Minutes,
		userInfo.RemainingTime.Seconds)
}
