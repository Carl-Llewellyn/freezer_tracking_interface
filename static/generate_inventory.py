import json
import random

# Configuration
NUM_ROOMS = 3
FREEZER_RANGE = (4, 7)      # Each room: 4 to 7 freezers
SHELVES_PER_FREEZER = 4
BOXES_PER_SHELF = (1, 8)   # Each shelf: 8 to 16 boxes
SAMPLES_PER_BOX = (3, 12)   # Each box: 3 to 12 samples

ROOM_NAMES = ["Room 1", "Room 2", "Room 3"]
BUILDINGS = ["A", "B", "C"]
FLOORS = ["1st Floor", "2nd Floor", "Basement"]

FREEZER_MODELS = ["Thermo ULT1786", "Panasonic MDF", "Fisher FZ-100", "Sanyo MDF"]
BOX_TYPES = ["Cardboard", "Plastic"]
SAMPLE_TYPES = ["eDNA", "Whole Fish"]
SPECIES = [
    "Salmo salar", "Oncorhynchus mykiss", "Esox lucius", "Gadus morhua",
    "Perca fluviatilis", "Cyprinus carpio", "Silurus glanis", "Lota lota"
]
COLLECTORS = ["Christoph Deeg", "Kristi Miller", "Carl Llewellyn", "Kyle Goff", "Art Bass", "Angela Schulze"]

def rand_label(prefix, n, shelf=None, freezer=None):
    if shelf is not None and freezer is not None:
        return f"{prefix}Box-F{freezer}S{shelf}-{n}"
    elif shelf is not None:
        return f"{prefix}Box-S{shelf}-{n}"
    return f"{prefix}Box-{n}"

def generate_inventory():
    inventory = {"rooms": []}
    for r in range(NUM_ROOMS):
        room_id = f"room{r+1}"
        room = {
            "id": room_id,
            "name": ROOM_NAMES[r],
            "metadata": {
                "location": f"Building {BUILDINGS[r]}, {FLOORS[r]}",
            },
            "freezers": []
        }
        num_freezers = random.randint(*FREEZER_RANGE)
        for f in range(num_freezers):
            freezer_id = f"freezer{r+1}-{f+1}"
            freezer = {
                "id": freezer_id,
                "name": f"Freezer #{f+1}",
                "metadata": {
                    "model": random.choice(FREEZER_MODELS),
                    "serial": f"{random.choice(['T', 'P', 'F', 'S'])}{r+1}{f+1:02d}",
                    "temperature": random.choice(["-60°C", "-70°C", "-80°C"])
                },
                "shelves": []
            }
            for s in range(SHELVES_PER_FREEZER):
                shelf_id = f"shelf{r+1}-{f+1}-{s+1}"
                shelf = {
                    "id": shelf_id,
                    "name": f"Shelf {s+1}",
                    "metadata": {"capacity": 20},
                    "boxes": []
                }
                num_boxes = random.randint(*BOXES_PER_SHELF)
                for b in range(num_boxes):
                    box_id = f"box{r+1}-{f+1}-{s+1}-{b+1}"
                    box_type = random.choice(BOX_TYPES)
                    box = {
                        "id": box_id,
                        "name": f"Box {b+1}",
                        "metadata": {
                            "box_type": box_type,
                            "label": rand_label(
                                chr(65+(b % 12)), b+1, shelf=s+1, freezer=f+1
                            )
                        },
                        "samples": []
                    }
                    num_samples = random.randint(*SAMPLES_PER_BOX)
                    for sm in range(num_samples):
                        is_edna = random.choice([True, False])
                        sample_type = SAMPLE_TYPES[0] if is_edna else SAMPLE_TYPES[1]
                        sample_id = f"sample{r+1}-{f+1}-{s+1}-{b+1}-{sm+1}"
                        sample_name = (
                            f"Sample EDNA-{r+1}{f+1}{s+1}{b+1}{sm+1:03d}"
                            if is_edna else
                            f"Fish-Body-{r+1}{f+1}{s+1}{b+1}{sm+1:03d}"
                        )
                        sample_meta = {
                            "collected_by": random.choice(COLLECTORS),
                            "date_collected": f"2025-04-{random.randint(1, 28):02d}"
                        }
                        if not is_edna:
                            sample_meta["species"] = random.choice(SPECIES)
                        sample = {
                            "id": sample_id,
                            "type": sample_type,
                            "name": sample_name,
                            "metadata": sample_meta
                        }
                        box["samples"].append(sample)
                    shelf["boxes"].append(box)
                freezer["shelves"].append(shelf)
            room["freezers"].append(freezer)
        inventory["rooms"].append(room)
    return inventory

if __name__ == "__main__":
    inventory = generate_inventory()
    with open("inventory.json", "w") as f:
        json.dump(inventory, f, indent=2)
    print("Generated inventory.json with random data!")