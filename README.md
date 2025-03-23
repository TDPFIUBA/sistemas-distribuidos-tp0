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
