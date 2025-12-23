package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

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

func printUsage() {
	fmt.Println("GoNauta - Cliente CLI para Nauta")
	fmt.Println("\nUso: gonauta <comando>")
	fmt.Println("Comandos disponibles:")
	fmt.Println("  login     - Guardar credenciales (usuario y contraseña)")
	fmt.Println("  connect   - Iniciar sesión en Nauta")
	fmt.Println("  logout    - Cerrar sesión activa")
	fmt.Println("  status    - Ver tiempo restante de la sesión activa")
	fmt.Println("  info      - Ver información completa del usuario")
	fmt.Println("  help      - Mostrar esta ayuda")
}

func handleLogin() {
	reader := bufio.NewReader(os.Stdin)

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

	if err := SaveCredentials(username, password); err != nil {
		fmt.Printf("Error guardando credenciales: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Credenciales guardadas exitosamente")
	fmt.Println("  Use 'gonauta connect' para iniciar sesión")
}

func handleConnect() {
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
	fmt.Println("\nUse 'gonauta status' para ver el tiempo restante")
	fmt.Println("Use 'gonauta logout' para cerrar la sesión")
}

func handleLogout() {
	sessionData, err := LoadSession()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
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
		fmt.Println("\nLa sesión puede haber expirado. Use 'gonauta connect' para iniciar sesión nuevamente")
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
