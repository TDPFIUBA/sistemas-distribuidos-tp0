# Sistemas Distribuidos - TP0

## Parte 1: Introducción a Docker

#### Dependencias

- Python >= 3.13

#### Ejercicios

<details>

<summary>Ejercicio 1</summary>

### Ejercicio N°1:

En este ejercicio creo un generador de configuraciones (YAML) para docker compose, donde se define:

- Un servicio de servidor
- N servicios de clientes (configurables)
- Una red compartida por todos los servicios

#### Dependencias

- PyYAML

#### Instalación

Instalar PyYAML

```bash
pip install pyyaml
```

Dar permisos para hacer el archivo ejectuable

```bash
chmod +x generar-compose.sh
```

#### Archivos agregados

- **generar-compose.py**: Script Python donde implementé la lógica para generar el archivo YAML
- **generar-compose.sh**: Script Bash solicitado por la consigna (ejecuta el subscript de Python)

#### Uso

```bash
./generar-compose.sh <archivo-salida> <n-clientes>
```

#### Ejemplos

Generar un archivo `docker-compose-dev.yaml` con `5` clientes:

```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

</details>

<details>

<summary>Ejercicio 2</summary>

### Ejercicio N°2

En este ejercicio mapeo los volúmenes del cliente y del servidor para poder modificar sus archivos de configuración sin necesidad de reconstruir las imágenes de Docker.

#### Mapeos de volúmenes:

- **Cliente:**
  ```yaml
  ./client/config.yaml:/config.yaml
  ```
- **Servidor:**
  ```yaml
  ./server/config.ini:/config.ini
  ```

Estos mapeos los implemente en el script `generar-compose.py`, asegurando que todos los YAML generados en el futuro los incluyan automáticamente.

#### Cambios adicionales:

- Eliminé variables de entorno del archivo `generar-compose.yml`, ya que ahora la configuración se realiza a través de los archivos `config.*`.
- Agregué un archivo `.dockerignore` tanto para el cliente como para el servidor. En estos agregué el archivo de configuración para que no se copie en el Dockerfile.

</details>

<details>

<summary>Ejercicio 3</summary>

### Ejercicio N°3

Cree el archivo `validar-echo-server.sh` que permite verificar el correcto funcionamiento del servidor.
Esto se hace mediante el comando `nc (netcat)`.

#### Respuestas según validación:

- **Exitosa:**
  ```
  action: test_echo_server | result: success
  ```
- **Error:**
  ```
  action: test_echo_server | result: fail
  ```

#### Cambios adicionales:

- Hice cambios en la validación de cantidad de clientes posibles en la generación del docker compose, ya que no se permitian 0 clientes previamente.

#### Uso

```bash
./validar-echo-server.sh
```

</details>

<details>

<summary>Ejercicio 4</summary>

### Ejercicio N°4

Agregué el manejo de la señal `SIGTERM` en el cliente como también en el servidor, haciendo que la aplicación termine de forma graceful.

#### Implementaciones

##### Servidor

Agregué el handler de la señal a la clase del servidor, donde al recibirla, este ejecuta el método `__handle_sigterm_signal` y libera los sockets.

```python
signal.signal(signal.SIGTERM, self.__handle_sigterm_signal)
```

##### Cliente

Agregué el handler de la señal al main file, donde se crea un channel que recibe señales.

```go
signalChannel := make(chan os.Signal, 1)
```

Este channel se utiliza para enviar señales `SIGTERM`

```go
signal.Notify(signalChannel, syscall.SIGTERM)
```

Se crea una goroutine para recibir las señales y cerrar las conexiones del cliente

```go
go func() {
  <-signalChannel
  log.Infof("action: sigterm_received | result: success | client_id: %v", clientConfig.ID)
  client.CloseConnection()
  os.Exit(0)
}()
```

</details>
