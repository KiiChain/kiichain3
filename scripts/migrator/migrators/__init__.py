from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Callable

class Migrator(ABC):
    @abstractmethod
    def migrate(self, data: dict):
        pass

    def delete_self(self) -> bool:
        return False
