import json
import sys

from migrators.root import Root

# Main starts the processing of the genesis
def main(input_genesis: str, output_genesis: str):
    # Load the main file
    with open(input_genesis, "r") as f:
        input_genesis_content = json.loads(f.read())

    # Process it
    input_genesis_content = Root().migrate(input_genesis_content)

    # Save it
    with open(output_genesis, "w") as f:
        f.write(json.dumps(input_genesis_content))

if __name__ == "__main__":
    if len(sys.argv) < 3:
        raise "Exactly two arguments are expected"

    # Take the input
    input_genesis = sys.argv[1]
    # Take the output
    output_genesis = sys.argv[2]

    # Execute the main function
    main(input_genesis, output_genesis)