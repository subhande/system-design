from typing import List
from abc import ABC, abstractmethod
from enum import Enum

# Elevator System Design

class ElevatorState(Enum):
    IDLE = 1
    MOVING_UP = 2
    MOVING_DOWN = 3

class Direction(Enum):
    UP = 1
    DOWN = 2

class DoorState(Enum):
    OPEN = 1
    CLOSED = 2


class Button(ABC):
    def __init__(self, status):
        self.__status = status

    def pressDown(self):
        pass

    @abstractmethod
    def isPressed(self):
        pass


class HallButton(Button):
    def __init__(self, buttonSign: Direction, sourceFloorNumber: int):
        self.__buttonSign = buttonSign

    def isPressed(self):
        pass


class ElevatorButton(Button):
    def __init__(self, destination_floor_number: int):
        self.__destination_floor_number = destination_floor_number

    def isPressed(self):
        pass



class ElevatorPanel:
    def __init__(self, floorButtons: List[ElevatorButton], openButton: ElevatorButton, closeButton: ElevatorButton):
        self.__floorButtons = floorButtons
        self.__openButton = openButton
        self.__closeButton = closeButton


class HallPanel:
    def __init__(self, up: HallButton, down: HallButton):
        self.__up = up
        self.__down = down


class Display:
    def __init__(self, floor: int, capacity: int, direction: Direction):
        self.__floor = floor
        self.__capacity = capacity
        self.__direction = direction

    def showElevatorDisplay(self):
        pass

    def showHallDisplay(self):
        pass

class Door:
    def __init__(self, state: DoorState):
        self.__state = state

    def isOpen(self):
        pass

class ElevatorCar:
    def __init__(self, id: int, door: Door, state: ElevatorState, display: Display, panel: ElevatorPanel, currentFloor: int):
        self.__id = id
        self.__door = door
        self.__state = state
        self.__display = display
        self.__panel = panel
        self.__currentFloor = currentFloor


    def openDoor(self):
        pass

    def closeDoor(self):
        pass

    def move(self):
        pass

    def stop(self):
        pass


class Floor:
    def __init__(self, displays: List[Display], panels: List[HallPanel]):
        self.__displays = displays
        self.__panels = panels

    def isBottomMost(self):
        pass

    def isTopMost(self):
        pass


class __ElevatorSystem(object):
    __instances = None

    def __new__(cls):
        if cls.__instances is None:
            cls.__instances = super(__ElevatorSystem, cls).__new__(cls)
        return cls.__instances

class ElevatorSystem(metaclass=__ElevatorSystem):
    def __init__(self, building):
        self.__building = building

    def monitoring(self):
        pass

    def dispatcher(self):
        pass



class __Building(object):
    __instances = None
    def __new__(cls):
        if cls.__instances is None:
            cls.__instances = super(__Building, cls).__new__(cls)
            return cls.__instances

class Building(metaclass=__Building):
  def __init__(self):
    self.__floor = []
    self.__elevator = []
