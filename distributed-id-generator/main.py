import time, random, os

# Initialize counter to 0
counter = 0

# Path to the file where the last saved ID is stored
PATH = "id.txt"

# Frequency to increment the counter
FREQUENCY = 1000


def restore_last_saved_id(offset):
    """
    Restore the last saved ID from the file and increment it by offset.
    If the file does not exist, return offset.
    """
    try:
        with open(PATH, "r") as f:
            return int(f.read()) + offset
    except FileNotFoundError:
        return 0 + offset


def generate_id(save_frequency):
    """
    Generate a unique ID based on the current time, a random machine ID, and a counter.
    Args:
        save_frequency: int: The frequency at which the counter should be saved to the file
    Returns:
        str: A unique ID
    """
    global counter
    # If counter is 0 and the file exists, restore the last saved ID
    if counter == 0 and os.path.exists(PATH):
        counter = restore_last_saved_id(offset=save_frequency)

    # Get the current time in milliseconds since epoch
    epoch_ms = int(time.time() * 1000)

    # Generate a random machine ID between 0 and 10, padded with zeros to 2 digits
    machine_id = str(random.randint(0, 10)).zfill(2)

    # Create the ID string
    id = f"{epoch_ms}{machine_id}{str(counter).zfill(4)}"

    # If the counter is a multiple of save_frequency, save the counter to the file
    if counter % save_frequency == 0:
        # Code to save the counter to the file will go here
        with open(PATH, "w") as f:
            f.write(str(counter))
    # Increment the counter
    counter += 1
    return id


if __name__ == "__main__":
    NO_OF_IDS = 1000000
    FREQUENCY = 100
    start = time.time()
    for i in range(NO_OF_IDS):
        generate_id(FREQUENCY)
    print(f"Generated {NO_OF_IDS} ids in {round((time.time() - start)* 1000, 2)} ms | Save Frequency: {FREQUENCY}")

    FREQUENCY = 1000
    start = time.time()
    for i in range(NO_OF_IDS):
        generate_id(FREQUENCY)
    print(f"Generated {NO_OF_IDS} ids in {round((time.time() - start)* 1000, 2)} ms | Save Frequency: {FREQUENCY}")

    FREQUENCY = 10000
    start = time.time()
    for i in range(NO_OF_IDS):
        generate_id(FREQUENCY)
    print(f"Generated {NO_OF_IDS} ids in {round((time.time() - start)* 1000, 2)} ms | Save Frequency: {FREQUENCY}")
