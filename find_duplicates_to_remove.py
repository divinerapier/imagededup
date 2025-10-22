import sys
import json

from pathlib import Path

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



if __name__ == '__main__':
    algorithm_name = sys.argv[1]
    dir = sys.argv[2]
    image_dir = Path(dir)
    algorithm = create_algorithm(name=algorithm_name)
    removes = algorithm.find_duplicates_to_remove(image_dir=image_dir)
    print(json.dumps(removes, indent=4))
