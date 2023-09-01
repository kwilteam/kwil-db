home_dir="/app/home_dir"

mkdir -p home_dir

# Initialize the node directory
/app/kwild utils init -o $home_dir


# check if home_dir/abci/config/genesis.json exists
genesis_path="$home_dir/abci/config/genesis.json"

if [ -e "$genesis_path" ]; then
    echo "File '$genesis_path' exists."
else
    echo "File '$genesis_path' does not exist."
    exit 1
fi


# Check if the config file exists
config_path="$home_dir/config.toml"

if [ -e "$config_path" ]; then
    echo "File '$config_path' exists."
else
    echo "File '$config_path' does not exist."
    exit 1
fi


