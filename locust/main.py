from locust import HttpUser, task, between
import random


class CarUser(HttpUser):
    @task
    def post(self):
        data = {
            "identifier": "car1",
            "lat": str(random.randint(30, 100)),
            "long": str(random.randint(30, 100)),
            "status": "running",
        }
        self.client.post('/', data=data)
    wait_time = between(1, 2)
