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

## Parte 2: Repaso de Comunicaciones

#### Dependencias

- Python >= 3.13

#### Ejercicios

<details>

<summary>Ejercicio 5</summary>

### Ejercicio N°5:

En este ejercicio implemente la comunicacion cliente-servidor, en la cual el cliente le envia al servidor los datos necesarios para realizar una apuesta, y el servidor lo procesa.

En primer lugar, defini la comunicación con el protocolo.

- Los datos para realizar una apuesta son: `NOMBRE`, `APELLIDO`, `DOCUMENTO`, `NACIMIENTO` y `NUMERO`.
- Estos datos se obtienen por medio de variables de entorno.

Para esto, se agregaron los valores DEFAULT de la consigna en el script que genera el docker compose:

```python
"environment": [f"CLI_ID={client_n}", "NOMBRE=Santiago Lionel","APELLIDO=Lorca","DOCUMENTO=30904465","NACIMIENTO=1999-03-17","NUMERO=7574"],
```

Los datos se envian del cliente hacia el servidor en el siguiente formato:

```bash
AGENCY=%s,FIRST_NAME=%s,LAST_NAME=%s,DOCUMENT=%s,BIRTHDATE=%s,NUMBER=%s\n
```

Una vez recibido del lado del servidor, la apuesta se guarda, y se responde con un mensaje:

```bash
RESULT=%s,MESSAGE=%s\n
```

Estos mensajes se reciben y se envian mediante sockets, teniendo en cuenta los short writes y short reads.

Para implementar esta logica, el `servidor` tiene una clase `Protocol` que maneja y encampsula esto:

```python
class Protocol
  def send_message(sock, data: bytes)
  def receive_message(sock)
```

```python
class ProtocolMessage
  def bytes_to_str(data: bytes)
  def str_to_bytes(data: str)
  def serialize_response(success: bool, message: str)
  def deserialize_bet(data: bytes)
```

Del lado del `cliente`, lo mismo, manejado con structs y funciones:

```go
type Protocol struct { // size=16 (0x10)
    conn net.Conn
}
func (p *Protocol) ReceiveMessage() (string, error)
func (p *Protocol) SendMessage(data []byte) error
```

```go
type BetMessage struct { // size=96 (0x60)
    Agency    string
    FirstName string
    LastName  string
    Document  string
    Birthdate string
    Number    string
}
func (m *BetMessage) Serialize() []byte
```

#### Ejemplo de uso:

Generar docker compose con un servidor y un cliente (con variables de entorno)

```bash
./generar-compose.sh docker-compose-dev.yaml 1
```

Levantar los servicios con Makefile

```bash
make docker-compose-up
```

En los logs se podran observar acciones del estilo:

- `action: apuesta_enviada`

  Generado cuando el cliente recibe la confirmación del servidor al enviar una apuesta

- `action: apuesta_almacenada`

  Generado cuando el servidor almacena la apuesta

</details>

<details>

<summary>Ejercicio 6</summary>

### Ejercicio N°6:

En este ejercicio implemente la comunicacion cliente-servidor, en la cual el cliente le envia al servidor los datos necesarios para realizar una apuesta o varias apuestas utilizando batches de estas.

En primer lugar, defini la comunicación con el protocolo (reutilizando lo generado para el ej5).

- Los datos para realizar una apuesta son: `NOMBRE`, `APELLIDO`, `DOCUMENTO`, `NACIMIENTO` y `NUMERO`.
- Estos datos se obtienen por medio de los registros de los archivos CSV.

#### Archivo de apuestas:

Para esto, se deben leer los datos del CSV asignado para el cliente, donde cada registro tiene una estructura:

```python
"Santiago Lionel,Lorca,30904465,1999-03-17,7574"
```

Este debe ser agregado en la carpeta `.data`, donde gracias a la generación de docker compose, se montara automaticamente.

```python
f"./.data/agency-{client_n}.csv:/data/agency.csv",
```

#### Comunicación:

Los datos se envian del cliente hacia el servidor en el siguiente formato:

```bash
BETS=%d;AGENCY=%s,FIRST_NAME=%s,LAST_NAME=%s,DOCUMENT=%s,BIRTHDATE=%s,NUMBER=%s;...
```

Una vez recibido del lado del servidor, la apuesta se guarda, y se responde con un mensaje:

```bash
RESULT=%s,MESSAGE=%s\n
```

Estos mensajes se reciben y se envian mediante sockets, teniendo en cuenta los short writes y short reads.

#### Codigo:

Para implementar esta logica, el `servidor` tiene una clase `Protocol` que maneja y encapsula esto:

```python
class Server
  def __process_bet_batch(self, bets: list[Bet])
```

- `__process_bet_batch`: Valida cantidades y guarda en caso exitoso

```python
class ProtocolMessage
  def deserialize_bets_batch(data: bytes)
```

- `deserialize_bets_batch`: Deserializa el batch, separandolo y utilizando el método `deserialize_bet` para cada una de ellas.

Del lado del `cliente`, lo mismo, manejado con structs y funciones:

```go
type BatchBetMessage struct { // size=24 (0x18)
    Bets []*BetMessage
}
func (m *BatchBetMessage) AddBet(bet *BetMessage)
func (m *BatchBetMessage) Serialize() []byte
```

- `AddBet`: Agrega una apuesta al listado de apuestas del batch.
- `Serialize`: Serializa el batch a bytes para el envio.

#### Variables de entorno

Se agrega `batch.maxAmount` y se eliminan `NOMBRE`, `APELLIDO`, `DOCUMENTO`, `NACIMIENTO` y `NUMERO`

#### Ejemplo de uso:

Agregar los archivos con apuestas para cada servidor. Donde `N` es el `ID` del cliente.

El formato de cada registro debe ser de la siguiente manera: `NOMBRE,APELLIDO,DOCUMENTO,NACIMIENTO,NUMERO`

```bash
.data/agency-{N}.csv
```

Generar docker compose con un servidor y N clientes

```bash
./generar-compose.sh docker-compose-dev.yaml N
```

Levantar los servicios con Makefile

```bash
make docker-compose-up
```

En los logs se podran observar acciones del estilo:

- `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`

  Generado cuando en el servidor todas las apuestas del batch fueron procesadas correctamente

- `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`

  Generado cuando en el servidor se detecta un error con alguna de las apuestas

</details>

<details>

<summary>Ejercicio 7</summary>

### Ejercicio N°7:

En este ejercicio modifique los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas (Esto sirve para que el servidor sepa cuando realizar el sorteo), y además puedan consultar sobre los ganadores del sorteo.

#### Variables de entorno

Se agrega `CLI_QTY` tanto al servidor como a los clientes. Esta variable contiene la cantidad de clientes generados con el script que genera el docker compose. Se utiliza para hacer validaciones del sorteo.

#### Comunicación:

Los mensajes que se envian del cliente hacia el servidor contienen uno de los siguientes formatos:

Envio de batch de apuestas

```bash
BETS=%d;AGENCY=%s,FIRST_NAME=%s,LAST_NAME=%s,DOCUMENT=%s,BIRTHDATE=%s,NUMBER=%s;...
```

Finalización de envio de apuestas

```bash
END,AGENCY=%s
```

Consulta de ganadores

```bash
WINNERS,AGENCY=%s
```

Una vez recibido del lado del servidor, se responde con un mensaje:

```bash
RESULT=%s,MESSAGE=%s
```

Estos mensajes se reciben y se envian mediante sockets, teniendo en cuenta los short writes y short reads.

#### Codigo:

Para implementar esta logica, se agrega en el `servidor`:

```python
class Server
  self._clients_qty = int(cli_qty)
  self._finished_clients = set()
  self._lottery_ran = False
  self._winners = {}

  def __handle_no_more_bets_msg(self, msg)
  def __handle_get_winner_msg(self, msg)
  def __run_lottery(self)
```

- `self._clients_qty`: Cantidad de clientes obtenida por variable de entorno. Utilizada para validar si se puede realizar la loteria.
- `self._finished_clients`: Set de clientes que terminaron de enviar sus apuestas.
- `self._lottery_ran`: Bool que indica si ya se ejecuto la loteria o no.
- `self._winners`: Dict para guardar los ganadores según cliente.

- `__handle_no_more_bets_msg`: Maneja el mensaje de que ya envio todas las apuestas de un cliente.
- `__handle_get_winner_msg`: Maneja el mensaje de consulta de ganadores de un cliente.
- `__run_lottery`: Ejecuta la loteria para obtener los ganadores.

```python
class ProtocolMessage
  def is_no_more_bets(data: bytes)
  def is_get_winner(data: bytes)
  def parse_no_more_bets(data: bytes)
  def parse_get_winner(data: bytes)
```

- `is_no_more_bets`: Valida si el mensaje es de fin de envio de apuestas.
- `is_get_winner`: Valida si el mensaje es de consulta de ganadores.
- `parse_no_more_bets`: Parsea el mensaje de fin de envio de apuestas.
- `parse_get_winner`: Parsea el mensaje de fin de consulta de ganadores.

Del lado del `cliente`, lo mismo, manejado con structs y funciones:

```go
type Client struct { // size=80 (0x50)
    config   ClientConfig
    conn     net.Conn
    protocol *Protocol
}

func (c *Client) GetWinners() (bool, error)
func (c *Client) GetWinnersLoop() error
func (c *Client) NotifyNoMoreBets() error
```

- `GetWinners`: Envia mensaje al servidor para obtener los ganadores.
- `GetWinnersLoop`: Loop, el cual ejecuta `GetWinnners` internamente tantas veces hasta que se envien los ganadores.
- `NotifyNoMoreBets`: Envia mensaje al servidor para notificar que no tiene más apuestas pendientes de envio.

```go
type NoMoreBetsMessage struct { // size=16 (0x10)
    AgencyID string
}
func (m *NoMoreBetsMessage) Serialize() []byte
```

- `Serialize`: Serializa el id de la agencia que termino de enviar apuestas.

```go
type GetWinnerMessage struct { // size=16 (0x10)
    AgencyID string
}
func (m *GetWinnerMessage) Serialize() []byte
```

- `Serialize`: Serializa el id de la agencia que consulta por los ganadores.

#### Ejemplo de uso:

Agregar los archivos con apuestas para cada servidor. Donde `N` es el `ID` del cliente.

El formato de cada registro debe ser de la siguiente manera: `NOMBRE,APELLIDO,DOCUMENTO,NACIMIENTO,NUMERO`

```bash
.data/agency-{N}.csv
```

Generar docker compose con un servidor y N clientes

```bash
./generar-compose.sh docker-compose-dev.yaml N
```

Levantar los servicios con Makefile

```bash
make docker-compose-up
```

En los logs se podran observar acciones del estilo:

- `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`

  Generado cuando el cliente consulta y obtiene los ganadores.

- `action: sorteo | result: success`

  Generado cuando en el servidor ejecuta la loteria.

</details>
