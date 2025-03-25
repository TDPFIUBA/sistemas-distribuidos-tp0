import logging
from common.utils import Bet
from common.protocol_message import ProtocolMessage


class Protocol:
    """Protocol for handling communication between client and server"""

    END_MESSAGE = ProtocolMessage.END_MESSAGE
    BUFFER_SIZE = 4

    @staticmethod
    def send_message(sock, data: bytes) -> int:
        """Send message"""
        data_qty_sent = 0
        data_qty = len(data)

        while data_qty_sent < data_qty:
            sent = sock.send(data[data_qty_sent:])
            if sent == 0:
                raise RuntimeError("Socket connection failed while sending message")
            data_qty_sent += sent

        return data_qty_sent

    @staticmethod
    def receive_message(sock) -> bytes:
        """Receive message until end is found"""
        data = b""
        while True:
            chunk = sock.recv(Protocol.BUFFER_SIZE)
            if not chunk:
                raise RuntimeError(
                    "Socket connection failed before receiving complete message"
                )

            data += chunk

            if Protocol.END_MESSAGE in chunk:
                break

        return data.rstrip(Protocol.END_MESSAGE)
