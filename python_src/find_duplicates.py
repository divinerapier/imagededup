import sys
import json

from pathlib import Path
from typing import Dict, List

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


class DuplicateItem:
    def __init__(self, filename: str, score: float):
        self.filename = filename
        self.score = score
    
    def to_dict(self):
        return {
            'filename': self.filename,
            'score': self.score
        }


class DuplicateResult:
    def __init__(self, filename: str, duplicate_list: List[DuplicateItem]):
        self.filename = filename
        self.duplicate_list = duplicate_list

    def add_duplicate(self, filename: str, score: float):
        self.duplicate_list.append(DuplicateItem(filename=filename, score=score))

    def to_dict(self):
        return {
            'filename': self.filename,
            'duplicate_list': [item.to_dict() for item in self.duplicate_list]
        }

    def to_json(self):
        return json.dumps(self.to_dict(), indent=4)


def find_duplicates(algorithm_name: str, image_dir: Path) -> Dict[str, DuplicateResult]:
    algorithm = create_algorithm(name=algorithm_name)
    duplicates_dict: Dict[str, DuplicateResult] = {}
    duplicates = algorithm.find_duplicates(image_dir=image_dir, scores=True)
    for filename, duplicate_list in duplicates.items():
        duplicate_result = DuplicateResult(filename=filename, duplicate_list=[])
        for duplicate in duplicate_list:
            duplicate_result.add_duplicate(filename=duplicate[0], score=float(duplicate[1]))
        duplicates_dict[filename] = duplicate_result
    
    # 将 DuplicateResult 对象转换为字典进行 JSON 序列化
    return {filename: result.to_dict() for filename, result in duplicates_dict.items()}



if __name__ == '__main__':
    algorithm_name = sys.argv[1]
    dir = sys.argv[2]
    image_dir = Path(dir)
    algorithm = create_algorithm(name=algorithm_name)
    duplicates_dict: Dict[str, DuplicateResult] = {}
    duplicates = algorithm.find_duplicates(image_dir=image_dir, scores=True)
    for filename, duplicate_list in duplicates.items():
        duplicate_result = DuplicateResult(filename=filename, duplicate_list=[])
        for duplicate in duplicate_list:
            duplicate_result.add_duplicate(filename=duplicate[0], score=float(duplicate[1]))
        duplicates_dict[filename] = duplicate_result
    
    # 将 DuplicateResult 对象转换为字典进行 JSON 序列化
    result_dict = {filename: result.to_dict() for filename, result in duplicates_dict.items()}
    print(json.dumps(result_dict, indent=4))
