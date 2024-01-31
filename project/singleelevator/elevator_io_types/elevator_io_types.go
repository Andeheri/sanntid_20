package elevatoriotypes

const N_FLOORS = 4
const N_BUTTONS = 3

type Dirn int
const ( 
    D_Down  Dirn = -1
    D_Stop  Dirn = 0
    D_Up Dirn = 1
)

type Button int
const (
    B_HallUp Button = 0
    B_HallDown Button = 1
    B_Cab Button = 2
)


typedef struct {
    int (*floorSensor)(void);
    int (*requestButton)(int, Button);
    int (*stopButton)(void);
    int (*obstruction)(void);
    
} ElevInputDevice;

typedef struct {
    void (*floorIndicator)(int);
    void (*requestButtonLight)(int, Button, int);
    void (*doorLight)(int);
    void (*stopButtonLight)(int);
    void (*motorDirection)(Dirn);
} ElevOutputDevice;
