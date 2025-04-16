from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Callable


class Migrator(ABC):
    @abstractmethod
    def migrate(self, data: dict):
        pass

    def delete_self(self) -> bool:
        return False

    def print_class_name(self):
        # Get the class name
        class_name = self.__class__.__name__

        # If the class name is NoChange we skip it
        if class_name == "NoChange":
            return
        
        print(f"Class {class_name} migrated successfully")
