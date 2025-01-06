# test_ghastly.py

import time
from client import GhastlyClient

def test_basic_operations():
    """Test basic database operations."""
    client = GhastlyClient()

    # Test health check
    assert client.health_check(), "Database should be healthy"
    print("✓ Health check passed")

    # Test put operation
    success = client.put("test_key", "test_value")
    assert success, "Put operation should succeed"
    print("✓ Put operation successful")

    # Test get operation
    value = client.get("test_key")
    assert value == "test_value", "Should retrieve the correct value"
    print("✓ Get operation successful")

    # Test search operation
    results = client.search("test_value")
    assert len(results) > 0, "Should find at least one result"
    assert results[0]['value'] == "test_value", "Should find the exact match"
    print("✓ Search operation successful")

    # Test delete operation
    success = client.delete("test_key")
    assert success, "Delete operation should succeed"
    print("✓ Delete operation successful")

    # Verify deletion
    value = client.get("test_key")
    assert value is None, "Value should be deleted"
    print("✓ Deletion verification successful")

def test_bulk_operations():
    """Test bulk operations."""
    client = GhastlyClient()

    # Test bulk put
    items = [
        ("key1", "The quick brown fox"),
        ("key2", "jumps over the lazy dog"),
        ("key3", "A quick brown dog jumps"),
    ]

    result = client.bulk_put(items)
    assert result['processed_count'] == len(items), "All items should be processed"
    print("✓ Bulk put operation successful")

    # Test semantic search
    results = client.search("quick animal jumping", limit=2)
    assert len(results) <= 2, "Should respect the limit parameter"
    print("✓ Semantic search successful")

    # Clean up
    for key, _ in items:
        client.delete(key)
    print("✓ Cleanup successful")

if __name__ == "__main__":
    try:
        print("Starting GhastlyDB client tests...")
        test_basic_operations()
        print("\nBasic operations tests passed!")

        time.sleep(1)  # Small pause between test sets

        test_bulk_operations()
        print("\nBulk operations tests passed!")

    except Exception as e:
        print(f"Tests failed: {e}")
    finally:
        print("\nTests completed.")