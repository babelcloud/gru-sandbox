.PHONY: clean clean-test clean-pyc clean-build docs help format lint test test-all coverage start install-dev-tools test-e2e fix-lint
.DEFAULT_GOAL := help

help: ## Show help message
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-dev-tools: ## Install development tools (black, isort, ruff, mypy)
	pip install black isort ruff mypy pytest pytest-cov

start: install-dev-tools ## Install all required dependencies and development tools
	pip install -r requirements.txt
	pip install -e .

clean: clean-build clean-pyc clean-test ## Remove all build, test, coverage and Python artifacts

clean-build: ## Remove build artifacts
	rm -fr build/
	rm -fr dist/
	rm -fr .eggs/
	find . -name '*.egg-info' -exec rm -fr {} +
	find . -name '*.egg' -exec rm -f {} +

clean-pyc: ## Remove Python file artifacts
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +
	find . -name '*~' -exec rm -f {} +
	find . -name '__pycache__' -exec rm -fr {} +

clean-test: ## Remove test and coverage artifacts
	rm -fr .tox/
	rm -f .coverage
	rm -fr htmlcov/
	rm -fr .pytest_cache/

format: install-dev-tools ## Format code using black and isort
	black gbox tests
	isort gbox tests

lint: install-dev-tools ## Check style
	ruff check gbox tests
	black --check gbox tests
	isort --check-only gbox tests
	mypy gbox

fix-lint: install-dev-tools ## Fix lint issues automatically using ruff
	ruff check --fix gbox tests --unsafe-fixes

test: ## Run tests
	pytest

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "Running Python SDK e2e tests..."
	pytest -v -s tests/e2e/

coverage: test ## Check code coverage
	pytest --cov=gbox
	coverage report -m
	coverage html
	open htmlcov/index.html 