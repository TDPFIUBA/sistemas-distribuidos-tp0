import socket
import logging
import signal
from common.protocol_message import ProtocolMessage
from common.protocol import Protocol
from common.utils import Bet, store_bets
from typing import Optional



class Server:
    def __init__(self, port, listen_backlog):
        """Server initialization"""
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)
        self._server_is_running = True

        self._client_socket = None
        self.__set_up_signal_handler()

    def __release_socket(self, release_socket):
        if release_socket:
            try:
                release_socket.shutdown(socket.SHUT_RDWR)
                release_socket.close()
            except Exception as e:
                logging.error("action: socket_release | result: fail | error: {e}")
        return None

    def __set_up_signal_handler(self):
        signal.signal(signal.SIGTERM, self.__handle_sigterm_signal)

    def __handle_sigterm_signal(self, signum, frame):
        """
        Handle SIGTERM signal

        When signal is received server stops accepting new connections and finishes the current
        connection with the client before closing the server.
        """
        logging.info("action: sigterm_signal | result: in_progress")
        try:
            self._server_is_running = False
            self._client_socket = self.__release_socket(self._client_socket)
            self._server_socket = self.__release_socket(self._server_socket)
            logging.info("action: sigterm_signal | result: success")
        except Exception as e:
            logging.error("action: sigterm_signal | result: fail | error: {e}")

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while self._server_is_running:
            self._client_socket = self.__accept_new_connection()
            if self._client_socket:
                self.__handle_client_connection()

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg = Protocol.receive_message(self._client_socket)
            addr = self._client_socket.getpeername()
            logging.info(
                f"action: receive_message | result: success | ip: {addr[0]}  | msg: {msg}"
            )

            bet = ProtocolMessage.deserialize_bet(msg)
            bet_saved, bet_msg = self.__process_bet(bet)

            response = ProtocolMessage.serialize_response(bet_saved, bet_msg)
            Protocol.send_message(self._client_socket, response)

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            try:
                error_response = ProtocolMessage.serialize_response(False, f"Error processing bet: {str(e)}")
                Protocol.send_message(self._client_socket, error_response)
            except:
                logging.error("action: send_error_response | result: fail")
        finally:
            self._client_socket = self.__release_socket(self._client_socket)

    def __process_bet(self, bet: Optional[Bet]):
        """
        Process a bet

        If the bet is valid, it is stored in the file system.
        If the bet is not valid, an error message is returned.
        """
        if not bet:
            return False, "Invalid bet data"

        store_bets([bet])
        logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
        return True, "Bet successfully registered"

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """
        logging.info("action: accept_connections | result: in_progress")
        try:
            if not self._server_is_running or self._server_socket is None:
                logging.info(
                    "action: accept_connections | result: fail | details: server not running"
                )
                return None
            c, addr = self._server_socket.accept()
            logging.info(
                f"action: accept_connections | result: success | ip: {addr[0]}"
            )
            return c
        except OSError as e:
            logging.error(f"action: accept_connections | result: fail | error: {e}")
            return None
