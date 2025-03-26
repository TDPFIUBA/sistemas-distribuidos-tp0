import socket
import logging
import signal
import multiprocessing
from multiprocessing import Process, Manager, Lock
from common.protocol_message import ProtocolMessage
from common.protocol import Protocol
from common.utils import Bet, store_bets, load_bets, has_won


class Server:
    def __init__(self, port, listen_backlog, cli_qty):
        """Server initialization"""
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)
        self._server_is_running = True

        self._clients_qty = int(cli_qty)
        
        manager = Manager()
        self._lock = Lock()
        self._finished_clients = manager.dict()
        self._lottery_ran = manager.Value('b', False)
        self._winners = manager.dict()
        
        self._process_pool = multiprocessing.Pool(processes=self._clients_qty)
        
        self.__set_up_signal_handler()

    def __release_socket(self, release_socket):
        if release_socket:
            try:
                release_socket.shutdown(socket.SHUT_RDWR)
                release_socket.close()
            except Exception as e:
                logging.error(f"action: socket_release | result: fail | error: {e}")
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
            self._server_socket = self.__release_socket(self._server_socket)
            self._process_pool.close()
            self._process_pool.join()
            logging.info("action: sigterm_signal | result: success")
        except Exception as e:
            logging.error(f"action: sigterm_signal | result: fail | error: {e}")

    def run(self):
        """
        Server main loop

        Accepts new connections and create a new process for each client connection
        """
        while self._server_is_running:
            try:
                client_socket = self.__accept_new_connection()
                if client_socket:
                    process = Process(target=self.__handle_client_connection, args=(client_socket,))
                    process.daemon = True
                    process.start()
            except Exception as e:
                logging.error(f"action: server_run | result: fail | error: {e}")
                if not self._server_is_running:
                    break

    def __handle_client_connection(self, client_socket):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg = Protocol.receive_message(client_socket)
            addr = client_socket.getpeername()
            logging.info(
                f"action: receive_message | result: success | ip: {addr[0]}  | msg: {msg}"
            )

            if ProtocolMessage.is_no_more_bets(msg):
                self.__handle_no_more_bets_msg(msg, client_socket)
            elif ProtocolMessage.is_get_winner(msg):
                self.__handle_get_winner_msg(msg, client_socket)
            else:
                bets = ProtocolMessage.deserialize_bets_batch(msg)
                best_saved, bets_msg = self.__process_bet_batch(bets)
                response = ProtocolMessage.serialize_response(best_saved, bets_msg)
                Protocol.send_message(client_socket, response)

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            try:
                error_response = ProtocolMessage.serialize_response(
                    False, f"Error processing req: {str(e)}"
                )
                Protocol.send_message(client_socket, error_response)
            except:
                logging.error("action: send_error_response | result: fail")
        finally:
            self.__release_socket(client_socket)

    def __handle_no_more_bets_msg(self, msg, client_socket):
        """Handle agency no more bets"""
        try:
            c_id = ProtocolMessage.parse_no_more_bets(msg)
            if not c_id:
                raise ValueError("Invalid handle_no_more_bets msg format")
            
            with self._lock:
                self._finished_clients[c_id] = True
                finished_count = len(self._finished_clients)
                should_run_lottery = finished_count == self._clients_qty and not self._lottery_ran.value
            
            logging.info(f"action: handle_no_more_bets | result: success | agency: {c_id} | total_agencies: {finished_count}")
            
            if should_run_lottery:
                self.__run_lottery()
            
            response = ProtocolMessage.serialize_response(True, "No more bets msg received")
            Protocol.send_message(client_socket, response)
            
        except Exception as e:
            logging.error(f"action: handle_no_more_bets | result: fail | error: {e}")
            raise

    def __handle_get_winner_msg(self, msg, client_socket):
        """Handle get winner"""
        try:
            c_id = ProtocolMessage.parse_get_winner(msg)
            if not c_id:
                raise ValueError("Invalid handle_get_winner msg format")
            
            with self._lock:
                lottery_ran = self._lottery_ran.value
                winners = list(self._winners.get(c_id, [])) if self._lottery_ran.value else []
            
            if not lottery_ran:
                response = ProtocolMessage.serialize_response(False, "0")
            else:
                response = ProtocolMessage.serialize_response(True, f"{len(winners)}")
                
            Protocol.send_message(client_socket, response)
            
        except Exception as e:
            logging.error(f"action: handle_get_winner | result: fail | error: {e}")
            raise

    def __run_lottery(self):
        """Run lottery when all agencies have reported"""
        try:
            with self._lock:
                if self._lottery_ran.value:
                    return
                
                winners = {}
                for bet in load_bets():
                    if has_won(bet):
                        c_id = str(bet.agency)
                        if c_id not in winners:
                            winners[c_id] = []
                        winners[c_id].append(bet.document)
                
                for key, value in winners.items():
                    self._winners[key] = value
                
                self._lottery_ran.value = True
                
            logging.info("action: sorteo | result: success")
                
        except Exception as e:
            logging.error(f"action: run_lottery | result: fail | error: {e}")
            raise

    def __process_bet_batch(self, bets: list[Bet]):
        """
        Process a batch of bets

        If all bets are valid, they are stored in the file system.
        If any bet is invalid, an error message is returned.
        """
        if not bets:
            logging.error("action: apuesta_recibida | result: fail | cantidad: 0")
            return False, "No valid bets in batch"

        bet_count = len(bets)

        for bet in bets:
            if not bet:
                logging.error(
                    f"action: apuesta_recibida | result: fail | cantidad: {bet_count}"
                )
                return False, "Invalid bet data in batch"

        store_bets(bets)

        logging.info(
            f"action: apuesta_recibida | result: success | cantidad: {bet_count}"
        )
        return True, f"Successfully processed {bet_count} bets"

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
