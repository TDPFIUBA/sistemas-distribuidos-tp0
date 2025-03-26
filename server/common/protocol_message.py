import logging
from common.utils import Bet


class ProtocolMessage:
    """Handles serialization and deserialization of protocol messages"""

    ENCODING = "utf-8"
    END_MESSAGE = b"\n"

    @staticmethod
    def bytes_to_str(data: bytes) -> str:
        """Converts bytes to string"""
        return data.decode(ProtocolMessage.ENCODING)

    @staticmethod
    def str_to_bytes(data: str) -> bytes:
        """Converts string to bytes"""
        return data.encode(ProtocolMessage.ENCODING)

    @staticmethod
    def serialize_response(success: bool, message: str) -> bytes:
        """Serializes a response message to be sent to the client"""
        result = "success" if success else "fail"
        response = f"RESULT={result},MESSAGE={message}"
        response = ProtocolMessage.str_to_bytes(response)
        return response + ProtocolMessage.END_MESSAGE

    @staticmethod
    def deserialize_bet(data: str) -> Bet:
        """Deserializes bet data from a string format KEY=VALUE,KEY=VALUE,..."""
        try:
            # Parse "KEY=VALUE,KEY=VALUE,..."
            bet_data = {}
            for pair in data.split(","):
                if "=" in pair:
                    key, value = pair.split("=", 1)
                    bet_data[key.lower()] = value

            bet = Bet(
                agency=bet_data.get("agency", ""),
                first_name=bet_data.get("first_name", ""),
                last_name=bet_data.get("last_name", ""),
                document=bet_data.get("document", ""),
                birthdate=bet_data.get("birthdate", ""),
                number=bet_data.get("number", ""),
            )
            return bet
        except Exception as e:
            logging.error(f"action: deserialize_bet | result: fail | error: {e}")
            return None

    @staticmethod
    def deserialize_bets_batch(data: bytes) -> list[Bet]:
        """Deserializes batch bet data"""
        try:
            data_str = ProtocolMessage.bytes_to_str(data)

            # Parse "BETS=N;BET1;BET2;..."
            split_data = data_str.split(";")
            bet_qty = split_data[0]

            try:
                bet_qty = int(bet_qty.split("=")[1])
            except:
                logging.error(
                    "action: deserialize_bets_batch | result: fail | error: invalid bet count"
                )
                return []

            # Parse "KEY=VALUE,KEY=VALUE,..."
            bets = []
            for i in range(1, len(split_data)):
                if i <= bet_qty:
                    bet_data = split_data[i]
                    bet = ProtocolMessage.deserialize_bet(bet_data)
                    if bet:
                        bets.append(bet)

            return bets
        except Exception as e:
            logging.error(f"action: deserialize_bets_batch | result: fail | error: {e}")
            return []
