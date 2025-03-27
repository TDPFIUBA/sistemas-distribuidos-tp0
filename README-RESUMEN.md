# Sistemas Distribuidos - TP0

## Introduccion

En este trabajo se implementa un servidor que atiende multiples clientes utilizando procesos. Cada cliente puede enviar varios tipos de mensajes para que el servidor finalmente pueda ejecutar la loteria y seleccionar los ganadores.

### Dependencias

- Python >= 3.13
- PyYAML _(se utiliza para el script que genera el docker compose dinamicamente)_

## Archivos

Cada cliente debe contar con un archivo `agency.csv` dentro de la carpeta `data`, en el cual los registros dentro de este deben tener el formato "Santiago Lionel,Lorca,30904465,1999-03-17,7574". Cada registro representa una apuesta _(NOMBRE,APELLIDO,DOCUMENTO,NACIMIENTO,NUMERO)_

- _En este trabajo al usar docker compose, se mapea el archivo ./.data/agency-{client_n}.csv al archivo agency.csv dentro del contenedor_

## Comunicación

### Mensajes

Cada cliente puede enviarle al servidor distintos tipos de mensajes para que el flujo de la loteria se concrete.

#### Cliente

##### Envio de apuestas en batch

Los batch de apuestas se envian en el siguiente formato. Cada apuesta ocupa aproximadamente 100 bytes

```
BETS=%d;AGENCY=%s,FIRST_NAME=%s,LAST_NAME=%s,DOCUMENT=%s,BIRTHDATE=%s,NUMBER=%s;...
```

##### Finalización de envio de apuestas

Mensaje que utiliza el cliente para avisar al servidor que no va a enviar más apuestas

```
END,AGENCY=%s
```

##### Consulta de ganadores

Mensaje para consultar al servidor quienes son los ganadores de la loteria

```
WINNERS,AGENCY=%s
```

#### Servidor

##### Respuesta

```
RESULT=%s,MESSAGE=%s
```

## Concurrencia

### Multiprocesamiento

Implemente un servidor que permite atender simultaneamnete a diferentes clientes, mientras garantiza que la consistencia de los datos compartidos se mantenga.

#### Consideraciones

Elegí multiprocesamiento en lugar de multihilo debido al Global Interpreter Lock (GIL) de Python, que limita la ejecución paralela de threads.

##### Procesos

- Se genera un `Pool` de procesos con la cantidad de clientes esperados
- Cada conexion nueva, se maneja en un proceso independiente (`multiprocessing.Process`). Cada uno configurado como daemon para cerrarse cuando el proceso principal termina
- `SIGTERM` cierra el `Pool` de procesos

##### Variables compartidas

- Para compartir los datos de la loteria entre procesos utilice `multiprocessing.Manager`
- Para sincronizar acceso a ellas, lo maneje con un `Lock`

##### Archivos

- Para manejar los accesos al archivo de `bets.csv` agregue un `Lock` para garantizar que solo un proceso acceda al archivo a la vez
