# ghastly_client.py

import grpc
from typing import List, Optional
import gen.ghastly_pb2 as pb2
import gen.ghastly_pb2_grpc as pb2_grpc

class GhastlyClient:
    """A client for interacting with GhastlyDB through gRPC."""

    def __init__(self, host: str = "localhost", port: int = 50051):
        """Initialize the GhastlyDB client.

        Args:
            host: The hostname where GhastlyDB is running
            port: The port number where GhastlyDB is listening
        """
        self.channel = grpc.insecure_channel(f"{host}:{port}")
        self.stub = pb2_grpc.GhastlyDBStub(self.channel)

    def put(self, key: str, value: str) -> bool:
        """Store a key-value pair in the database.

        Args:
            key: The key to store
            value: The value to store

        Returns:
            bool: True if successful, False otherwise

        Raises:
            grpc.RpcError: If the gRPC call fails
        """
        request = pb2.PutRequest(key=key, value=value)
        try:
            response = self.stub.Put(request)
            return response.success
        except grpc.RpcError as e:
            print(f"Error storing key-value pair: {e}")
            raise

    def get(self, key: str) -> Optional[str]:
        """Retrieve a value by its key.

        Args:
            key: The key to look up

        Returns:
            Optional[str]: The value if found, None otherwise

        Raises:
            grpc.RpcError: If the gRPC call fails
        """
        request = pb2.GetRequest(key=key)
        try:
            response = self.stub.Get(request)
            return response.value if response.found else None
        except grpc.RpcError as e:
            print(f"Error retrieving value: {e}")
            raise

    def delete(self, key: str) -> bool:
        """Delete a key-value pair from the database.

        Args:
            key: The key to delete

        Returns:
            bool: True if successful, False otherwise

        Raises:
            grpc.RpcError: If the gRPC call fails
        """
        request = pb2.DeleteRequest(key=key)
        try:
            response = self.stub.Delete(request)
            return response.success
        except grpc.RpcError as e:
            print(f"Error deleting key: {e}")
            raise

    def search(self, query: str, limit: int = 10, score_threshold: float = 0.0) -> List[dict]:
        """Search for similar vectors in the database.

        Args:
            query: The search query
            limit: Maximum number of results to return
            score_threshold: Minimum similarity score threshold

        Returns:
            List[dict]: List of search results, each containing key, value, and score

        Raises:
            grpc.RpcError: If the gRPC call fails
        """
        request = pb2.SearchRequest(
            query=query,
            limit=limit,
            score_threshold=score_threshold
        )
        try:
            response = self.stub.Search(request)
            return [
                {
                    'key': result.key,
                    'value': result.value,
                    'score': result.score
                }
                for result in response.results
            ]
        except grpc.RpcError as e:
            print(f"Error performing search: {e}")
            raise

    def bulk_put(self, items: List[tuple]) -> dict:
        """Store multiple key-value pairs in the database.

        Args:
            items: List of (key, value) tuples to store

        Returns:
            dict: Statistics about the bulk operation

        Raises:
            grpc.RpcError: If the gRPC call fails
        """
        try:
            def generate_requests():
                for key, value in items:
                    yield pb2.PutRequest(key=key, value=value)

            response = self.stub.BulkPut(generate_requests())
            return {
                'processed_count': response.processed_count,
                'failed_keys': response.failed_keys
            }
        except grpc.RpcError as e:
            print(f"Error performing bulk put: {e}")
            raise

    def health_check(self) -> bool:
        """Check if the database is healthy and serving requests.

        Returns:
            bool: True if healthy, False otherwise

        Raises:
            grpc.RpcError: If the gRPC call fails
        """
        request = pb2.HealthCheckRequest()
        try:
            response = self.stub.HealthCheck(request)
            return response.status == pb2.HealthCheckResponse.SERVING
        except grpc.RpcError as e:
            print(f"Error checking health: {e}")
            raise

    def close(self):
        """Close the gRPC channel."""
        self.channel.close()