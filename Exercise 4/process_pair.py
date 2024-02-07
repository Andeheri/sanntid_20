
import socket
import os
import time


def main():
    ip_address = "localhost"
    primary_port = 44834
    timeout = 2
    buffer_size = 10
    number = -1

    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.settimeout(timeout)
        try:
            s.connect((ip_address, primary_port))
            # Process is a backup
            print("Connected to primary")
            print("Backup started")
            while True:
                recieved_bytes = s.recv(buffer_size)
                if len(recieved_bytes) == 0:  # Recieved 0-buffer
                    raise Exception
                number = int.from_bytes(recieved_bytes, byteorder='little')
        except Exception as e:
            print(e)
    # Process is a primary
    print("Primary started")

    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        s.settimeout(timeout)
        try:
            s.bind((ip_address, primary_port))
            s.listen()
            # Start up backup
            os.system(f'gnome-terminal -- python3 {__file__}')

            conn, addr = s.accept()
            with conn:
                print("Connected to backup")
                while True:
                    number += 1
                    print(number)
                    conn.sendall(number.to_bytes(buffer_size, byteorder='little'))
                    time.sleep(0.5)
        except Exception as e:
            print(e)
    return


if __name__ == '__main__':
    main()