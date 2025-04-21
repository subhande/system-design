from enum import Enum
from abc import ABC, abstractmethod



# Enumerations and custom data type
class PaymentStatus(Enum):
    COMPLETED = 1 
    FAILED = 2
    PENDING = 3
    UNPAID = 4
    REFUNDED = 5
    
class AccountStatus(Enum):
    ACTIVE = 1
    CLOSED = 2
    CANCELED = 3 
    BLACKLISTED = 4
    NONE = 5
    
# Custom Person data type class
class Person:
  def __init__(self, name, address, phone, email):
    self.__name = name
    self.__address = address
    self.__phone = phone
    self.__email = email


# Custom Address data type class
class Address:
  def __init__(self, zip_code, address, city, state, country):
    self.__zip_code = zip_code
    self.__address = address
    self.__city = city
    self.__state = state
    self.__country = country
    
#####################################
######   ParkingSpot Class   #######
#####################################

# ParkingSpot is an abstract class
class ParkingSpot(ABC):
  def __init__(self, id, isFree, vehicle):
    self.__id = id
    self.__isFree = isFree
    self.__vehicle = vehicle # Refers to an instance of the Vehicle class

  def get_is_free(self):
    pass

  # vehicle here refers to an instance of the Vehicle class
  @abstractmethod
  def assign_vehicle(self, vehicle):
    pass

  def remove_vehicle(self):
    pass

class Handicapped(ParkingSpot):
  # vehicle here refers to an instance of the Vehicle class
  def __init__(self, id, isFree, vehicle):
    super().__init__(id, isFree, vehicle)

  # vehicle here refers to an instance of the Vehicle class
  def assign_vehicle(self, vehicle):
    pass

class Compact(ParkingSpot):
  # vehicle here refers to an instance of the Vehicle class
  def __init__(self, id, isFree, vehicle):
    super().__init__(id, isFree, vehicle)

  # vehicle here refers to an instance of the Vehicle class
  def assign_vehicle(self, vehicle):
    pass

class Large(ParkingSpot):
  # vehicle here refers to an instance of the Vehicle class
  def __init__(self, id, isFree, vehicle):
    super().__init__(id, isFree, vehicle)

  # vehicle here refers to an instance of the Vehicle class
  def assign_vehicle(self, vehicle):
    pass

class Motorcycle(ParkingSpot):
  # vehicle here refers to an instance of the Vehicle class
  def __init__(self, id, isFree, vehicle):
    super().__init__(id, isFree, vehicle)

  # vehicle here refers to an instance of the Vehicle class
  def assign_vehicle(self, vehicle):
    pass

#########################################
######   Vehicle Class   ################
#########################################


# Vehicle is an abstract class
class Vehicle(ABC):
  def __init__(self, license_no):
    self.__license_no = license_no

  # ticket here refers to an instance of the ParkingTicket class
  @abstractmethod
  def assign_ticket(self, ticket):
    pass

class Car(Vehicle):
  # ticket here refers to an instance of the ParkingTicket class
  def __init__(self, license_no):
    super().__init__(license_no)
  
  # ticket here refers to an instance of the ParkingTicket class
  def assign_ticket(self, ticket):
    pass

class Van(Vehicle):
  # ticket here refers to an instance of the ParkingTicket class
  def __init__(self, license_no):
    super().__init__(license_no)

  # ticket here refers to an instance of the ParkingTicket class
  def assign_ticket(self, ticket):
    pass

class Truck(Vehicle):
  # ticket here refers to an instance of the ParkingTicket class
  def __init__(self, license_no, ticket):
    super().__init__(license_no, ticket)

  # ticket here refers to an instance of the ParkingTicket class
  def assign_ticket(self, ticket):
    pass

class MotorCycle(Vehicle):
  # ticket here refers to an instance of the ParkingTicket class
  def __init__(self, license_no, ticket):
    super().__init__(license_no, ticket)

  # ticket here refers to an instance of the ParkingTicket class
  def assign_ticket(self, ticket):
    pass


#########################################
######   Account Class   ##########
#########################################


class Account(ABC):
  # Data members
  def __init__(self, user_name, password, person, status):
    self.__user_name = user_name
    self.__password = password
    self.__person = person # Refers to an instance of the Person class
    self.__status = status # Refers to the AccountStatus enum

  @abstractmethod
  def reset_password(self):
    pass

class Admin(Account):
  def __init__(self, user_name, password, person, status):
    super().__init__(user_name, password, person, status)

  # spot here refers to an instance of the ParkingSpot class
  def add_parking_spot(self, spot):
    pass

  # display_board here refers to an instance of the DisplayBoard class
  def add_display_board(self, display_board):
    pass

  # entrance here refers to an instance of the Entrance class
  def add_entrance(self, entrance):
    pass

  # exit here refers to an instance of the Exit class
  def add_exit(self, exit):
    pass

  def reset_password(self):
    # Will implement the functionality in this class
    pass

class ParkingAttendant(Account):
  def __init__(self, user_name, password, person, status):
    super().__init__(user_name, password, person, status)

  def process_ticket(self, ticket_number):
    pass

  def reset_password(self):
    # Will implement the functionality in this class
    pass



#########################################
######   DisplayBoard Class   ##########
#########################################



class DisplayBoard:
  def __init__(self, id):
        self.__id = id
        self.__parking_spots = {}

  # Member functions
  def add_parking_spot(self, spot_type, spots):
    pass
  def show_free_slot(self):
    pass

class ParkinRate:
  def __init__(self, hours, rate):
    self.__hours = hours
    self.__rate = rate

  # Member function
  def calculate(self):
    pass

#########################################
######   Entrance and Exit Class   #####
#########################################

class Entrance:
  def __init__(self, id, ticket):
    self.__id = id

  # ticket here refers to an instance of the ParkingTicket class
  def get_ticket(self):
    pass

class Exit:
  def __init__(self, id, ticket):
    self.__id = id

  # ticket here refers to an instance of the ParkingTicket class
  def validate_ticket(self, ticket):
    # Perform validation logic for the parking ticket
    # Calculate parking charges, if necessary
    # Handle the exit process
    pass


##########################################
######   ParkingTicket Class   ##########
##########################################


class ParkingTicket:
  def __init__(self, ticket_no, timestamp, exit, amount, status, vehicle, payment, entrance, exit_ins):
    self.__ticket_no = ticket_no
    self.__timestamp = timestamp
    self.__exit = exit
    self.__amount = amount
    self.__status = status
    
    # Following are the instances of their respective classes
    self__vehicle = vehicle
    self__payment = payment
    self__entrance = entrance
    self__exit_ins = exit_ins
    
    
#################################################
######   Payment Class   ######################
#################################################



# Payment is an abstract class
class Payment(ABC):
  def __init__(self, amount, status, timestamp):
    self.__amount = amount
    self.__status = status # Refers to the PaymentStatus enum
    self.__timestamp = timestamp

  @abstractmethod
  def initiate_transaction(self):
    pass

class Cash(Payment):
  def __init__(self, amount, status, timestamp):
    super().__init__(amount, status, timestamp)

  def initiate_transaction(self):
    pass

class CreditCard(Payment):
  def __init__(self, amount, status, timestamp):
    super().__init__(amount, status, timestamp)

  def initiate_transaction(self):
    pass


##############################################
######   ParkingLot Class   #################
##############################################

# The __ParkingLot is a singleton class that ensures it will have only one active instance at a time
# Both the Entrance and Exit classes use this class to create and close parking tickets
class __ParkingLot(object):
  __instances = None
  
  def __new__(cls):
    if cls.__instances is None:
        cls.__instances = super(__ParkingLot, cls).__new__(cls)
    return cls.__instances

class ParkingLot(metaclass=__ParkingLot):
  def __init__(self, id, name, address, parking_rate):
    # Call the name, address and parking_rate 
    self.__id = id
    self.__name = name
    self.__address = address
    self.__parking_rate = parking_rate

    # Create initial entrance and exit hashmaps respectively
    self.__entrance = {}
    self.__exit = {}

    # Create a hashmap that identifies all currently generated tickets using their ticket number
    self.__tickets = {}

  # entrance here refers to an instance of the Entrance class
  def add_entrance(self, entrance):
    pass

  # exit here refers to an instance of the Exit class
  def add_exit(self, exit):
    pass

  # This function allows parking tickets to be available at multiple entrances
  # vehicle here refers to an Vehicle of the Exit class
  # Returns a ParkingTicket instance
  def get_parking_ticket(self, vehicle):
    pass

   # type here refers to an instance of the ParkingSpot class
  def is_full(self, type):
    pass