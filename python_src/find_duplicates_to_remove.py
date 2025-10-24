import sys
import json

from pathlib import Path
from typing import List

from imagededup.methods import PHash, CNN, DHash, AHash, WHash


def create_algorithm(name: str):
    if name == 'phasher':
        return PHash()
    elif name == 'cnn':
        return CNN()
    elif name == 'dhash':
        return DHash()
    elif name == 'ahash':
        return AHash()
    elif name == 'whash':
        return WHash()
    else:
        raise ValueError(f'Invalid algorithm: {name}')


def find_duplicates_to_remove(algorithm_name: str, image_dir: Path) -> List[str]:
    algorithm = create_algorithm(name=algorithm_name)
    removes = algorithm.find_duplicates_to_remove(image_dir=image_dir)
    return removes


if __name__ == '__main__':
    algorithm_name = sys.argv[1]
    dir = sys.argv[2]
    image_dir = Path(dir)
    algorithm = create_algorithm(name=algorithm_name)
    removes = algorithm.find_duplicates_to_remove(image_dir=image_dir)
    print(json.dumps(removes, indent=4))
