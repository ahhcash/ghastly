from setuptools import setup, find_packages

setup(
    name="ghastly-client",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "grpcio>=1.69.0",
        "protobuf>=5.29.0",
    ],
    python_requires=">=3.7",
)