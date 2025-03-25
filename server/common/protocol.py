import logging
from common.utils import Bet

class Protocol:
    """Protocol for handling communication between client and server"""
    
    END_MESSAGE = b'\n'
    BUFFER_SIZE = 4
    ENCODING = 'utf-8'

    @staticmethod
    def bytes_to_str(data):
        """Converts bytes to string"""
        return data.decode(Protocol.ENCODING)

    @staticmethod
    def str_to_bytes(data):
        """Converts string to bytes"""
        return data.encode(Protocol.ENCODING)
    
    @staticmethod
    def serialize_response(success, message):
        """Serializes a response message to be sent to the client"""
        result = "success" if success else "fail"
        response = f"RESULT={result},MESSAGE={message}"
        response = Protocol.str_to_bytes(response)
        return response + Protocol.END_MESSAGE
    
    @staticmethod
    def deserialize_bet(data):
        """Deserializes bet data from a bytes/string format KEY=VALUE,KEY=VALUE,..."""
        try:
            data = Protocol.bytes_to_str(data)

            # Parse "KEY=VALUE,KEY=VALUE,..."
            bet_data = {}
            for pair in data.split(','):
                if '=' in pair:
                    key, value = pair.split('=', 1)
                    bet_data[key.lower()] = value
            
            bet = Bet(
                agency=bet_data.get('agency', ''),
                first_name=bet_data.get('first_name', ''),
                last_name=bet_data.get('last_name', ''),
                document=bet_data.get('document', ''),
                birthdate=bet_data.get('birthdate', ''),
                number=bet_data.get('number', '')
            )
            return bet
        except Exception as e:
            logging.error(f"action: deserialize_bet | result: fail | error: {e}")
            return None
    
    @staticmethod
    def send_message(sock, data):
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
    def receive_message(sock):
        """Receive message until end is found"""
        data = b''
        while True:
            chunk = sock.recv(Protocol.BUFFER_SIZE)
            if not chunk:
                raise RuntimeError("Socket connection failed before receiving complete message")
            
            data += chunk
            
            if Protocol.END_MESSAGE in chunk:
                break
                
        return data.rstrip(Protocol.END_MESSAGE)
