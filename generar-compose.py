import yaml
import sys

COMPOSE_NAME = "tp0"

DEFAULT_EXTRA_CONF = {}
DEFAULT_CLIENTS_QTY = 1

ARGV_QTY = 2
ARGV_FILE_IDX = 1
ARGV_CLIENTS_QTY_IDX = 2

SYS_ERROR_CODE = 1


# Used to increase indent in yaml (pyyaml)
class Dumper(yaml.Dumper):
    def increase_indent(self, flow=False, *args, **kwargs):
        return super().increase_indent(flow=flow, indentless=False)


def generate_networks_config():
    return {
        "testing_net": {
            "ipam": {"driver": "default", "config": [{"subnet": "172.25.125.0/24"}]}
        }
    }


def generate_server_config(extra_conf=DEFAULT_EXTRA_CONF):
    return {
        "container_name": "server",
        "image": "server:latest",
        "entrypoint": "python3 /main.py",
        "environment": ["PYTHONUNBUFFERED=1"],
        **extra_conf,
    }


def generate_client_config(client_n, extra_conf=DEFAULT_EXTRA_CONF):
    return {
        "container_name": f"client{client_n}",
        "image": "client:latest",
        "entrypoint": "/client",
        "environment": [f"CLI_ID={client_n}"],
        **extra_conf,
    }


def generate_clients_config(
    clients_qty=DEFAULT_CLIENTS_QTY, extra_conf=DEFAULT_EXTRA_CONF
):
    return {
        f"client{n}": generate_client_config(n, extra_conf)
        for n in range(1, clients_qty + 1)
    }


def generate_yaml_config(clients_qty=DEFAULT_CLIENTS_QTY):
    networks = generate_networks_config()
    system_networks = {"networks": list(networks.keys())}
    return {
        "name": COMPOSE_NAME,
        "services": {
            "server": generate_server_config(system_networks),
            **generate_clients_config(
                clients_qty, {**system_networks, "depends_on": ["server"]}
            ),
        },
        "networks": networks,
    }


def print_usage():
    message = f"""
    Usage: {sys.argv[0]} <output-file> <n-clients>
        * output-file: file path to save docker compose data
        * n-clients: N° of clients to generate (optional, default: {DEFAULT_CLIENTS_QTY})
    """
    print(message)


def handle_command_error():
    print_usage()
    sys.exit(SYS_ERROR_CODE)


def validate_argv_qty():
    if len(sys.argv) < ARGV_QTY:
        handle_command_error()


def validate_clients_qty():
    try:
        clients_qty = int(sys.argv[ARGV_CLIENTS_QTY_IDX])
        if clients_qty <= 0:
            raise Exception
        return clients_qty
    except IndexError:
        return DEFAULT_CLIENTS_QTY
    except Exception:
        handle_command_error()


def save_file(output_file, yaml_config):
    try:
        with open(output_file, "w") as f:
            yaml.dump(
                yaml_config,
                f,
                default_flow_style=False,
                sort_keys=False,
                Dumper=Dumper,
            )
        print(f"File created successfully: {output_file}")
    except IOError as e:
        print(f"Error ({output_file}): {e}")
        sys.exit(SYS_ERROR_CODE)


def main():
    validate_argv_qty()

    output_file = sys.argv[ARGV_FILE_IDX]
    clients_qty = validate_clients_qty()
    yaml_config = generate_yaml_config(clients_qty)

    save_file(output_file, yaml_config)


if __name__ == "__main__":
    main()
