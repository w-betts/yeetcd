#!/usr/bin/env python3
"""
Tests for agent configurations in opencode.json.

These tests verify that:
- spec agent is configured with orchestrator prompt and subagent delegation
- vibe agent is configured with full permissions
- build and plan agents use OpenCode defaults (no custom config)
- All required agent configurations are present
"""

import json
import unittest
from pathlib import Path


class TestOpenCodeJsonStructure(unittest.TestCase):
    """Test cases for opencode.json file structure."""

    def setUp(self):
        """Set up test fixtures."""
        self.config_path = Path(__file__).parent / "opencode.json"
        self.config = None
        if self.config_path.exists():
            with open(self.config_path, 'r') as f:
                self.config = json.load(f)

    def test_config_file_exists(self):
        """Verify opencode.json configuration file exists."""
        self.assertTrue(
            self.config_path.exists(),
            f"opencode.json not found at {self.config_path}"
        )

    def test_config_is_valid_json(self):
        """Verify opencode.json is valid JSON."""
        if not self.config_path.exists():
            self.skipTest("opencode.json does not exist")
        
        try:
            with open(self.config_path, 'r') as f:
                config = json.load(f)
            self.assertIsInstance(config, dict)
        except json.JSONDecodeError as e:
            self.fail(f"opencode.json is not valid JSON: {e}")

    def test_config_has_agent_section(self):
        """Verify opencode.json has 'agent' configuration section."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        self.assertIn(
            "agent",
            self.config,
            "opencode.json should have 'agent' configuration section"
        )

    def test_config_has_schema(self):
        """Verify opencode.json has $schema field."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        self.assertIn(
            "$schema",
            self.config,
            "opencode.json should have $schema field for validation"
        )


class TestSpecAgentConfiguration(unittest.TestCase):
    """Test cases for spec agent configuration."""

    def setUp(self):
        """Set up test fixtures."""
        self.config_path = Path(__file__).parent / "opencode.json"
        self.config = None
        if self.config_path.exists():
            with open(self.config_path, 'r') as f:
                self.config = json.load(f)

    def test_spec_agent_exists(self):
        """Verify spec agent is configured in opencode.json."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        self.assertIn(
            "spec",
            agents,
            "spec agent should be configured in opencode.json"
        )

    def test_spec_agent_has_orchestrator_mode(self):
        """Verify spec agent is configured as primary/orchestrator mode."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "spec" not in agents:
            self.skipTest("spec agent not configured")
        
        spec_config = agents["spec"]
        self.assertEqual(
            spec_config.get("mode"),
            "primary",
            "spec agent should have mode='primary' for orchestration"
        )

    def test_spec_agent_has_orchestrator_prompt(self):
        """Verify spec agent uses orchestrator prompt."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "spec" not in agents:
            self.skipTest("spec agent not configured")
        
        spec_config = agents["spec"]
        prompt = spec_config.get("prompt", "")
        
        # Should reference orchestrator prompt file
        self.assertIn(
            "orchestrator",
            prompt.lower(),
            "spec agent should use orchestrator prompt"
        )

    def test_spec_agent_can_delegate_to_planner(self):
        """Verify spec agent can delegate to planner subagent."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "spec" not in agents:
            self.skipTest("spec agent not configured")
        
        spec_config = agents["spec"]
        permissions = spec_config.get("permission", {})
        task_permissions = permissions.get("task", {})
        
        # Should allow delegating to planner
        self.assertIn(
            "planner",
            task_permissions,
            "spec agent should be able to delegate to planner"
        )
        self.assertEqual(
            task_permissions.get("planner"),
            "allow",
            "spec agent should have 'allow' permission for planner"
        )

    def test_spec_agent_can_delegate_to_test_writer(self):
        """Verify spec agent can delegate to test-writer subagent."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "spec" not in agents:
            self.skipTest("spec agent not configured")
        
        spec_config = agents["spec"]
        permissions = spec_config.get("permission", {})
        task_permissions = permissions.get("task", {})
        
        self.assertIn(
            "test-writer",
            task_permissions,
            "spec agent should be able to delegate to test-writer"
        )
        self.assertEqual(
            task_permissions.get("test-writer"),
            "allow",
            "spec agent should have 'allow' permission for test-writer"
        )

    def test_spec_agent_can_delegate_to_implementer(self):
        """Verify spec agent can delegate to implementer subagent."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "spec" not in agents:
            self.skipTest("spec agent not configured")
        
        spec_config = agents["spec"]
        permissions = spec_config.get("permission", {})
        task_permissions = permissions.get("task", {})
        
        self.assertIn(
            "implementer",
            task_permissions,
            "spec agent should be able to delegate to implementer"
        )
        self.assertEqual(
            task_permissions.get("implementer"),
            "allow",
            "spec agent should have 'allow' permission for implementer"
        )


class TestVibeAgentConfiguration(unittest.TestCase):
    """Test cases for vibe agent configuration."""

    def setUp(self):
        """Set up test fixtures."""
        self.config_path = Path(__file__).parent / "opencode.json"
        self.config = None
        if self.config_path.exists():
            with open(self.config_path, 'r') as f:
                self.config = json.load(f)

    def test_vibe_agent_exists(self):
        """Verify vibe agent is configured in opencode.json."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        self.assertIn(
            "vibe",
            agents,
            "vibe agent should be configured in opencode.json"
        )

    def test_vibe_agent_has_primary_mode(self):
        """Verify vibe agent is configured as primary mode."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "vibe" not in agents:
            self.skipTest("vibe agent not configured")
        
        vibe_config = agents["vibe"]
        self.assertEqual(
            vibe_config.get("mode"),
            "primary",
            "vibe agent should have mode='primary'"
        )

    def test_vibe_agent_has_full_permissions(self):
        """Verify vibe agent has full tool access for direct implementation."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "vibe" not in agents:
            self.skipTest("vibe agent not configured")
        
        vibe_config = agents["vibe"]
        permissions = vibe_config.get("permission", {})
        
        # Should have broad edit permissions
        edit_perm = permissions.get("edit", {})
        self.assertTrue(
            edit_perm == "allow" or edit_perm.get("*") == "allow",
            "vibe agent should have full edit permissions"
        )
        
        # Should have broad bash permissions
        bash_perm = permissions.get("bash", {})
        self.assertTrue(
            bash_perm == "allow" or bash_perm.get("*") == "allow",
            "vibe agent should have full bash permissions"
        )

    def test_vibe_agent_can_optionally_delegate(self):
        """Verify vibe agent can optionally delegate to subagents."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "vibe" not in agents:
            self.skipTest("vibe agent not configured")
        
        vibe_config = agents["vibe"]
        permissions = vibe_config.get("permission", {})
        task_permissions = permissions.get("task", {})
        
        # vibe agent should be able to delegate to subagents if needed
        # Check if task permissions exist and allow delegation
        if task_permissions:
            # If task permissions are defined, they should allow subagents
            for subagent in ["planner", "test-writer", "implementer"]:
                if subagent in task_permissions:
                    self.assertEqual(
                        task_permissions.get(subagent),
                        "allow",
                        f"vibe agent should be able to delegate to {subagent}"
                    )


class TestBuildPlanAgentDefaults(unittest.TestCase):
    """Test cases for build and plan agent default configurations."""

    def setUp(self):
        """Set up test fixtures."""
        self.config_path = Path(__file__).parent / "opencode.json"
        self.config = None
        if self.config_path.exists():
            with open(self.config_path, 'r') as f:
                self.config = json.load(f)

    def test_build_agent_uses_opencode_defaults(self):
        """Verify build agent uses OpenCode defaults (minimal custom config)."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        
        # According to the plan, build and plan should use OpenCode defaults
        # This means they might not be in the config at all, or have minimal config
        if "build" in agents:
            build_config = agents["build"]
            # If build is defined, it should be minimal (just mode, no custom prompt)
            self.assertNotIn(
                "prompt",
                build_config,
                "build agent should use OpenCode defaults without custom prompt"
            )

    def test_plan_agent_uses_opencode_defaults(self):
        """Verify plan agent uses OpenCode defaults (minimal custom config)."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        
        # According to the plan, build and plan should use OpenCode defaults
        if "plan" in agents:
            plan_config = agents["plan"]
            # If plan is defined, it should be minimal (just mode, no custom prompt)
            self.assertNotIn(
                "prompt",
                plan_config,
                "plan agent should use OpenCode defaults without custom prompt"
            )


class TestSubagentConfigurations(unittest.TestCase):
    """Test cases for subagent configurations."""

    def setUp(self):
        """Set up test fixtures."""
        self.config_path = Path(__file__).parent / "opencode.json"
        self.config = None
        if self.config_path.exists():
            with open(self.config_path, 'r') as f:
                self.config = json.load(f)

    def test_planner_subagent_exists(self):
        """Verify planner subagent is configured."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        self.assertIn(
            "planner",
            agents,
            "planner subagent should be configured"
        )

    def test_planner_is_subagent_mode(self):
        """Verify planner is configured as subagent mode."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "planner" not in agents:
            self.skipTest("planner not configured")
        
        planner_config = agents["planner"]
        self.assertEqual(
            planner_config.get("mode"),
            "subagent",
            "planner should have mode='subagent'"
        )

    def test_test_writer_subagent_exists(self):
        """Verify test-writer subagent is configured."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        self.assertIn(
            "test-writer",
            agents,
            "test-writer subagent should be configured"
        )

    def test_test_writer_is_subagent_mode(self):
        """Verify test-writer is configured as subagent mode."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "test-writer" not in agents:
            self.skipTest("test-writer not configured")
        
        test_writer_config = agents["test-writer"]
        self.assertEqual(
            test_writer_config.get("mode"),
            "subagent",
            "test-writer should have mode='subagent'"
        )

    def test_implementer_subagent_exists(self):
        """Verify implementer subagent is configured."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        self.assertIn(
            "implementer",
            agents,
            "implementer subagent should be configured"
        )

    def test_implementer_is_subagent_mode(self):
        """Verify implementer is configured as subagent mode."""
        if self.config is None:
            self.skipTest("opencode.json does not exist or is invalid")
        
        agents = self.config.get("agent", {})
        if "implementer" not in agents:
            self.skipTest("implementer not configured")
        
        implementer_config = agents["implementer"]
        self.assertEqual(
            implementer_config.get("mode"),
            "subagent",
            "implementer should have mode='subagent'"
        )


if __name__ == '__main__':
    unittest.main()
