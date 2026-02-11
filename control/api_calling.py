import requests

def call_api(url, method, data=None, token=None):
    headers = {
        "Content-Type": "application/json"
    }

    if token:
        headers["Authorization"] = f"Bearer {token}"

    response = requests.request(method, url, headers=headers, json=data)
    return response.json()

def delete_all_grinds():
    url = "http://localhost:8080/api/v1/grinds/delete-all"
    method = "DELETE"
    data = None
    response = call_api(url, method, data)
    return response


if __name__ == "__main__":
    MENU = """
    1. Delete all grinds
    2. Create a grind
    3. Get all grinds
    4. Get a grind
    """

    print(MENU)
    api_choice = input("Enter the API choice: ")
    if api_choice == "1":
        response = delete_all_grinds()
        print(response)
    else:
        print("Invalid choice")
        print(MENU)