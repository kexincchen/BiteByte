import asyncio
import aiohttp
import json
from datetime import datetime


def load_config():
    try:
        with open('config.json', 'r') as f:
            return json.load(f)
    except FileNotFoundError:
        print("Error: config.json not found")
        exit(1)
    except json.JSONDecodeError:
        print("Error: Invalid JSON in config.json")
        exit(1)


async def create_order(session, order_data, config):
    order_data["merchant_id"] = config["merchant_id"]

    try:
        async with session.post(f"{config['api_base_url']}/orders", json=order_data) as response:
            if response.status == 201:
                result = await response.json()
                print(f"Customer {order_data['customer_id']} created order successfully")
                print(f"Order details: {json.dumps(order_data, indent=2)}")
                return result
            else:
                error_text = await response.text()
                print(
                    f"Failed to create order for customer {order_data['customer_id']}. Status: {response.status}, "
                    f"Error: {error_text}")
                return None
    except Exception as e:
        print(f"Error creating order for customer {order_data['customer_id']}: {e}")
        return None


async def main():
    """Main function: Create concurrent orders based on config"""
    config = load_config()

    async with aiohttp.ClientSession() as session:
        # Create tasks for all orders
        tasks = [create_order(session, order_data, config) for order_data in config["orders"]]

        start_time = datetime.now()
        results = await asyncio.gather(*tasks)
        end_time = datetime.now()

        # Calculate statistics
        successful_orders = sum(1 for order in results if order is not None)
        total_attempts = len(config["orders"])
        duration = (end_time - start_time).total_seconds()

        print("\n=== Test Results ===")
        print(f"Duration: {duration:.2f} seconds")
        print(f"Successful orders: {successful_orders}/{total_attempts}")
        print(f"Success rate: {(successful_orders / total_attempts) * 100:.2f}%")


if __name__ == "__main__":
    asyncio.run(main())
