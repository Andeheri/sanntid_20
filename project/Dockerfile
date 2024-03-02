#build container with 
#   docker build -t elevator .

#run container with 
#   docker run -it --rm elevator

# To run with broadcast
# sudo docker network create --driver bridge --subnet 192.168.0.0/24 --gateway 192.168.0.1 --opt "com.docker.network.bridge.enable_ip_masquerade"="true" --opt "com.docker.network.bridge.enable_icc"="true" --opt "com.docker.network.bridge.host_binding_ipv4"="0.0.0.0" mynetwork
# sudo docker run --network=mynetwork -it --rm elevator

FROM golang:1.20

WORKDIR /app

ENV NAME=test_elevator

#install tmux for terminal magic
RUN apt-get update && apt-get install -y tmux

#create a startup script to run both the simulator and the app side by side
COPY <<EOF ./start.sh
#!/bin/sh
tmux new-session -d -s "$NAME"
tmux send-keys -t "$NAME" "./app" Enter
tmux split-window -h -t "$NAME"
tmux send-keys -t "$NAME" "./simulator" Enter
tmux attach-session -t "$NAME"
EOF
#remove CR endings and set the file as executable
RUN sed -i 's/\r$//' ./start.sh  && \  
    chmod +x ./start.sh

# Add the elevator simulator
ADD https://github.com/TTK4145/Simulator-v2/releases/download/v1.5/SimElevatorServer ./simulator
RUN chmod +x simulator

# Add the source code
COPY . .
RUN go build -o app main.go

CMD [ "./start.sh" ]