# build container with 
# sudo docker build -t elevator .

# run container with 
# sudo docker run -it --rm elevator

FROM golang:1.20

WORKDIR /app

#install tmux for terminal magic
RUN apt-get update && apt-get install -y tmux

#create a startup script to run both the simulator and the app side by side
COPY <<EOF ./start.sh
#!/bin/sh
tmux new-session -d 
tmux send-keys "sleep 0.1 && ./app" Enter
tmux split-window -h
tmux send-keys "./simulator" Enter
tmux rename-window $(hostname -i)
tmux attach-session
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