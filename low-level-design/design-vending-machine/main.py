# Vending Machine Low Level Design

from abc import ABC, abstractmethod
from enum import Enum
from collections import defaultdict
from threading import Lock

class Coin(Enum):
    PENNY = 0.01
    NICKEL = 0.05
    DIME = 0.10
    QUARTER = 0.25

class Note(Enum):
    ONE = 1.00
    FIVE = 5.00
    TEN = 10.00
    TWENTY = 20.00
    FIFTY = 50.00
    HUNDRED = 100.00

class ProductType(Enum):
    CHOCOLATE = 1
    SNACK = 2
    BEVERAGE = 3
    OTHER = 4

class Product:
    def __init__(self, pid: int, name: str, price: float, ptype: ProductType):
        self.pid = pid
        self.name = name
        self.price = price
        self.ptype = ptype

    def __repr__(self):
        return f"Product(name={self.name}, price={self.price})"


class Inventory:
    def __init__(self):
        self.products = defaultdict(int)

    def add_product(self, product: Product, quantity: int):
        self.products[product.pid] += quantity

    def remove_product(self, product: Product):
        if self.products[product.pid] > 0:
            self.products[product.pid] -= 1
            if self.products[product.pid] == 0:
                del self.products[product.pid]
        else:
            print("Product not available in inventory")

    def update_quantity(self, product: Product, quantity: int):
        if product.pid in self.products:
            self.products[product.pid] = quantity
        else:
            print("Product not found in inventory")

    def get_product_count(self, product: Product) -> int:
        return self.products[product.pid]

    def is_available(self, product: Product) -> bool:
        return product.pid in self.products and self.products[product.pid] > 0


class VendingMachineState(ABC):
    def __init__(self, vending_machine):
        self.vending_machine = vending_machine

    @abstractmethod
    def select_product(self, product: Product):
        pass

    @abstractmethod
    def insert_coin(self, coin: Coin):
        pass

    @abstractmethod
    def insert_note(self, note: Note):
        pass

    @abstractmethod
    def dispense_product(self):
        pass

    @abstractmethod
    def return_change(self):
        pass


class IdleState(VendingMachineState):
    def __init__(self, vending_machine):
        self.vending_machine = vending_machine

    def select_product(self, product: Product):
        if self.vending_machine.inventory.is_available(product):
            self.vending_machine.selected_product = product
            self.vending_machine.set_state(self.vending_machine.ready_state)
            print(f"Product selected: {product.name} for ${product.price:.2f}")
        else:
            print(f"Product not available: {product.name}")

    def insert_coin(self, coin: Coin):
        print("Please select a product first.")

    def insert_note(self, note: Note):
        print("Please select a product first.")

    def dispense_product(self):
        print("Please select a product and make payment.")

    def return_change(self):
        print("No change to return.")


class ReadyState(VendingMachineState):
    def __init__(self, vending_machine):
        self.vending_machine = vending_machine

    def select_product(self, product: Product):
        print("Product already selected. Please make payment.")

    def insert_coin(self, coin: Coin):
        self.vending_machine.add_coin(coin)
        print(f"Coin inserted: {coin.name}")
        self.check_payment_status()

    def insert_note(self, note: Note):
        self.vending_machine.add_note(note)
        print(f"Note inserted: {note.name}")
        self.check_payment_status()

    def dispense_product(self):
        print("Please make payment first.")

    def return_change(self):
        change = self.vending_machine.total_payment
        if change > 0:
            print(f"Change returned: ${change:.2f}")
            self.vending_machine.reset_payment()
        else:
            print("No change to return.")
        self.vending_machine.reset_selected_product()
        self.vending_machine.set_state(self.vending_machine.idle_state)

    def check_payment_status(self):
        if self.vending_machine.total_payment >= self.vending_machine.selected_product.price:
            self.vending_machine.set_state(self.vending_machine.dispense_state)
            print("Payment successful. Dispensing product...")

class DispenseState(VendingMachineState):
    def __init__(self, vending_machine):
        self.vending_machine = vending_machine

    def select_product(self, product: Product):
        print("Product already selected. Please collect the dispensed product.")

    def insert_coin(self, coin: Coin):
        print("Payment already made. Please collect the dispensed product.")

    def insert_note(self, note: Note):
        print("Payment already made. Please collect the dispensed product.")

    def dispense_product(self):
        self.vending_machine.set_state(self.vending_machine.ready_state)
        product = self.vending_machine.selected_product
        self.vending_machine.inventory.remove_product(product)
        print(f"Product dispensed: {product.name}")
        self.vending_machine.set_state(self.vending_machine.return_change_state)

    def return_change(self):
        print("Please collect the dispensed product first.")

class ReturnChangeState(VendingMachineState):
    def __init__(self, vending_machine):
        self.vending_machine = vending_machine

    def select_product(self, product: Product):
        print("Please collect the change first.")

    def insert_coin(self, coin: Coin):
        print("Please collect the change first.")

    def insert_note(self, note: Note):
        print("Please collect the change first.")

    def dispense_product(self):
        print("Product already dispensed. Please collect the change.")

    def return_change(self):
        change = self.vending_machine.total_payment - self.vending_machine.selected_product.price
        if change > 0:
            print(f"Change returned: ${change:.2f}")
            self.vending_machine.reset_payment()
        else:
            print("No change to return.")
        self.vending_machine.reset_selected_product()
        self.vending_machine.set_state(self.vending_machine.idle_state)


class VendingMachine:
    _instance = None
    _lock = Lock()

    def __new__(cls):
        with cls._lock:
            if cls._instance is None:
                cls._instance = super().__new__(cls)
                cls._instance.inventory = Inventory()
                cls._instance.idle_state = IdleState(cls._instance)
                cls._instance.ready_state = ReadyState(cls._instance)
                cls._instance.dispense_state = DispenseState(cls._instance)
                cls._instance.return_change_state = ReturnChangeState(cls._instance)
                cls._instance.current_state = cls._instance.idle_state
                cls._instance.selected_product = None
                cls._instance.total_payment = 0.0
        return cls._instance

    @classmethod
    def get_instance(cls):
        return cls()

    def select_product(self, product: Product):
        self.current_state.select_product(product)

    def insert_coin(self, coin: Coin):
        self.current_state.insert_coin(coin)

    def insert_note(self, note: Note):
        self.current_state.insert_note(note)

    def dispense_product(self):
        self.current_state.dispense_product()

    def return_change(self):
        self.current_state.return_change()

    def set_state(self, state: VendingMachineState):
        self.current_state = state
        print(f"State changed to: {self.current_state.__class__.__name__}")

    def add_coin(self, coin: Coin):
        self.total_payment += coin.value
        print(f"Total payment updated: ${self.total_payment:.2f}")

    def add_note(self, note: Note):
        self.total_payment += note.value
        print(f"Total payment updated: ${self.total_payment:.2f}")

    def reset_payment(self):
        self.total_payment = 0.0
        print("Payment reset.")

    def reset_selected_product(self):
        self.selected_product = None
        print("Selected product reset.")


class VendingMachineDemo:
    @staticmethod
    def run():
        vending_machine = VendingMachine.get_instance()

        # Add products to the inventory
        coke = Product(1, "Coke", 1.5, ProductType.BEVERAGE)
        pepsi = Product(2, "Pepsi", 1.5, ProductType.BEVERAGE)
        water = Product(3, "Water", 1.0, ProductType.BEVERAGE)

        vending_machine.inventory.add_product(coke, 5)
        vending_machine.inventory.add_product(pepsi, 3)
        vending_machine.inventory.add_product(water, 2)

        # Select a product
        vending_machine.select_product(coke)

        # Insert coins
        vending_machine.insert_coin(Coin.QUARTER)
        vending_machine.insert_coin(Coin.QUARTER)
        vending_machine.insert_coin(Coin.QUARTER)
        vending_machine.insert_coin(Coin.QUARTER)

        # Insert a note
        vending_machine.insert_note(Note.FIVE)

        # Dispense the product
        vending_machine.dispense_product()

        # Return change
        vending_machine.return_change()

        # Select another product
        vending_machine.select_product(pepsi)

        # Insert insufficient payment
        vending_machine.insert_coin(Coin.QUARTER)

        # Try to dispense the product
        vending_machine.dispense_product()

        # Insert more coins
        vending_machine.insert_coin(Coin.QUARTER)
        vending_machine.insert_coin(Coin.QUARTER)
        vending_machine.insert_coin(Coin.QUARTER)
        vending_machine.insert_coin(Coin.QUARTER)

        # Dispense the product
        vending_machine.dispense_product()

        # Return change
        vending_machine.return_change()

if __name__ == "__main__":
    VendingMachineDemo.run()
