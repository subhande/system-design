# Consistent Hashing

from abc import ABC, abstractmethod
import hashlib
from bisect import bisect, bisect_right, bisect_left
# Basic Hashing

class StorageNode:
    def __init__(self, name: str, host: str):
        self.name = name
        self.host = host

    def __repr__(self):
        return f"StorageNode(name={self.name}, host={self.host})"

    def putFile(self, path: str):
        # Simulate storing a file in the storage node
        print(f"Storing {path} in {self.name}")

    def getFile(self, path: str):
        # Simulate retrieving a file from the storage node
        print(f"Retrieving {path} from {self.name}")



class Hashing(ABC):
    def __init__(self, n: int):
        self.storageNodes = [
            StorageNode(name=f"A{i}", host=f"storage-{i}.com") for i in range(n)
        ]
        self.noOfNodes = n
    @abstractmethod
    def hashFunc(self, key: str) -> int:
        pass

    @abstractmethod
    def uploadFile(self, fileName: str):
        pass

    @abstractmethod
    def downloadFile(self, fileName: str):
        pass


class BasicHashing(Hashing):
    """
    This class implements a basic hashing mechanism to store
    and retrieve files in a distributed storage system.
    """
    def __init__(self, n: int):
        self.storageNodes = [
            StorageNode(name=f"A{i}", host=f"storage-{i}.com") for i in range(n)
        ]
        self.noOfNodes = n


    def hashFunc(self, key: str) -> int:
        """
        The function sums the bytes present in the `key` and then
        take a mod with 5. This hash function thus generates output
        in the range [0, 4].
        """
        return sum(bytearray(key.encode('utf-8'))) % self.noOfNodes


    def uploadFile(self, fileName: str):
        """
        The function takes a file name as input and stores it in
        the storage node which is determined by the hash function.
        """
        # The hash function is called to determine the index of the
        index = self.hashFunc(fileName)
        # the storage node is determined by the hash function.
        node = self.storageNodes[index]
        # The file is stored in the storage node.
        return node.putFile(fileName)


    def downloadFile(self, fileName: str):
        """
        The function takes a file name as input and retrieves it
        from the storage node which is determined by the hash function.
        """
        # The hash function is called to determine the index of the
        index = self.hashFunc(fileName)
        # the storage node is determined by the hash function.
        node = self.storageNodes[index]
        # The file is retrieved from the storage node.
        return node.getFile(fileName)


class ConsistentHashing(Hashing):
    """
    This class implements a consistent hashing mechanism to store
    and retrieve files in a distributed storage system.
    """

    def __init__(self, totalSlots: int):
        """
        The constructor initializes the storage nodes and the
        number of nodes.
        """
        self.storageNodes = []
        self.totalSlots = totalSlots
        self._keys = []


    def hashFunc(self, key: str) -> int:
        """
        hashFunc creates an integer equivalent of a SHA256 hash and
        takes a modulo with the total number of slots in hash space.
        """
        hsh = hashlib.sha256()

        # converting data into bytes and passing it to hash function
        hsh.update(bytes(key.encode('utf-8')))

        # converting the HEX digest into equivalent integer value
        return int(hsh.hexdigest(), 16) % self.totalSlots

    def addStorageNode(self, node: StorageNode):
        """addStorageNode function adds a new node in the system and returns the key
        from the hash space where it was placed
        """

        # handling error when hash space is full.
        if len(self._keys) == self.totalSlots:
            raise Exception("hash space is full")

        key = self.hashFunc(node.host)

        # find the index where the key should be inserted in the keys array
        # this will be the index where the Storage Node will be added in the
        # nodes array.
        index = bisect(self._keys, key)

        # if we have already seen the key i.e. node already is present
        # for the same key, we raise Collision Exception
        if index > 0 and self._keys[index - 1] == key:
            raise Exception("collision occurred")

        # Perform data migration

        # insert the node_id and the key at the same `index` location.
        # this insertion will keep nodes and keys sorted w.r.t keys.
        self.storageNodes.insert(index, node)
        self._keys.insert(index, key)

        return key

    def removeStorageNode(self, node: StorageNode):
        """
        removeStorageNode removes the node and returns the key
        from the hash space on which the node was placed.
        """

        # handling error when space is empty
        if len(self._keys) == 0:
            raise Exception("hash space is empty")

        key = self.hashFunc(node.host)

        # we find the index where the key would reside in the keys
        index = bisect_left(self._keys, key)

        # if key does not exist in the array we raise Exception
        if index >= len(self._keys) or self._keys[index] != key:
            raise Exception("node does not exist")

        # Perform data migration

        # now that all sanity checks are done we popping the
        # keys and nodes at the index and thus removing the presence of the node.
        self._keys.pop(index)
        self.storageNodes.pop(index)

        return key


    def getNode(self, item: str) -> StorageNode:
        """
        getNode function returns the node which is responsible
        for the item.
        """

        key = self.hashFunc(item)

        # handling error when space is empty
        if len(self._keys) == 0:
            raise Exception("hash space is empty")

        # we find the first node to the right of this key
        # if bisect_right returns index which is out of bounds then
        # we circle back to the first in the array in a circular fashion.
        index = bisect_right(self._keys, key) % len(self._keys)

        return self.storageNodes[index]


    def uploadFile(self, fileName: str):
        """
        The function takes a file name as input and stores it in
        the storage node which is determined by the hash function.
        """

        # Get the node responsible for the file
        node = self.getNode(fileName)

        # The file is stored in the storage node.
        return node.putFile(fileName)

    def downloadFile(self, fileName: str):
        """
        The function takes a file name as input and retrieves it
        from the storage node which is determined by the hash function.
        """

        # Get the node responsible for the file
        node = self.getNode(fileName)

        # The file is retrieved from the storage node.
        return node.getFile(fileName)






if __name__ == "__main__":

    print("=" * 50)

    # Create an instance of BasicHashing
    basic_hashing = BasicHashing(5)

    # Upload files
    basic_hashing.uploadFile("file1.txt")
    basic_hashing.uploadFile("file2.txt")
    # Download files
    basic_hashing.downloadFile("file1.txt")
    basic_hashing.downloadFile("file2.txt")


    # Upload files
    basic_hashing.uploadFile("file1.txt")
    basic_hashing.uploadFile("file2.txt")
    # Download files
    basic_hashing.downloadFile("file1.txt")
    basic_hashing.downloadFile("file2.txt")

    print("=" * 50)

    # Create an instance of BasicHashing
    basic_hashing = BasicHashing(3)

    # Upload files
    basic_hashing.uploadFile("file1.txt")
    basic_hashing.uploadFile("file2.txt")
    # Download files
    basic_hashing.downloadFile("file1.txt")
    basic_hashing.downloadFile("file2.txt")


    # Upload files
    basic_hashing.uploadFile("file1.txt")
    basic_hashing.uploadFile("file2.txt")
    # Download files
    basic_hashing.downloadFile("file1.txt")
    basic_hashing.downloadFile("file2.txt")



    print("=" * 50)

    consisitent_hashing = ConsistentHashing(100)

    # add storage nodes
    for i in range(10):
        node = StorageNode(name=f"Node-{i}", host=f"node-{i}.com")
        consisitent_hashing.addStorageNode(node)

    # print(consisitent_hashing.storageNodes)
    print(consisitent_hashing._keys)

    # Upload files
    consisitent_hashing.uploadFile("file1.txt")
    consisitent_hashing.uploadFile("file10.txt")
    # Download files
    consisitent_hashing.downloadFile("file1.txt")
    consisitent_hashing.downloadFile("file10.txt")

    # add storage nodes
    for i in range(10, 15):
        node = StorageNode(name=f"Node-{i}", host=f"node-{i}.com")
        consisitent_hashing.addStorageNode(node)

    # print(consisitent_hashing.storageNodes)
    print(consisitent_hashing._keys)

    # Upload files
    consisitent_hashing.uploadFile("file1.txt")
    consisitent_hashing.uploadFile("file10.txt")
    # Download files
    consisitent_hashing.downloadFile("file1.txt")
    consisitent_hashing.downloadFile("file10.txt")
