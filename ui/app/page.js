"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  AlertIcon,
  Badge,
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Checkbox,
  Collapse,
  Divider,
  Flex,
  FormControl,
  FormLabel,
  Heading,
  HStack,
  IconButton,
  Input,
  Select,
  SimpleGrid,
  Stack,
  Stat,
  StatLabel,
  StatNumber,
  Switch,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
  Textarea,
  useDisclosure,
  VStack,
} from "@chakra-ui/react";
import { ChevronDownIcon, ChevronUpIcon } from "@chakra-ui/icons";

const getFallbackApiBase = () => {
  if (typeof window === "undefined") return "";
  const host = window.location.hostname;
  if (host === "localhost" || host === "127.0.0.1") {
    return "http://localhost:8080";
  }
  return `${window.location.origin}/api`;
};

export default function Page() {
  const [sessionId, setSessionId] = useState("");
  const [apiBase, setApiBase] = useState(getFallbackApiBase);
  const [aiEnabled, setAiEnabled] = useState(false);
  const [chatLog, setChatLog] = useState([]);
  const [chatInput, setChatInput] = useState("");
  const [chatAgent, setChatAgent] = useState("auto");
  const [busy, setBusy] = useState(false);
  const [errors, setErrors] = useState([]);
  const [snapshot, setSnapshot] = useState(null);
  const [loadingSnapshot, setLoadingSnapshot] = useState(false);
  const [activeSection, setActiveSection] = useState("jobs");
  const [searchText, setSearchText] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [pagination, setPagination] = useState({});

  const append = (role, text) => {
    setChatLog((prev) => [...prev, { role, text }]);
  };

  useEffect(() => {
    let ignore = false;
    const fetchRuntimeConfig = async () => {
      try {
        const resp = await fetch("/runtime-config", { cache: "no-store" });
        if (!resp.ok) return;
        const data = await resp.json();
        if (ignore) return;
        if (data.apiBase) setApiBase(data.apiBase);
        if (typeof data.aiEnabled === "boolean") setAiEnabled(data.aiEnabled);
      } catch (err) {
      }
    };
    fetchRuntimeConfig();
    return () => {
      ignore = true;
    };
  }, []);

  useEffect(() => {
    if (!apiBase) return;
    let ignore = false;
    const fetchMeta = async () => {
      try {
        const resp = await fetch(`${apiBase}/meta`);
        const data = await resp.json();
        if (!resp.ok || typeof data.ai_enabled !== "boolean") return;
        if (!ignore) setAiEnabled(data.ai_enabled);
      } catch (err) {
      }
    };
    fetchMeta();
    return () => {
      ignore = true;
    };
  }, [apiBase]);

  const postJSON = async (path, payload, method = "POST") => {
    if (!apiBase) {
      throw new Error("API base not set");
    }
    const resp = await fetch(`${apiBase}${path}`, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    let data = {};
    try {
      data = await resp.json();
    } catch (err) {
    }
    if (!resp.ok) {
      throw new Error(data.error || "Request failed");
    }
    return data;
  };

  const sendMessage = async (message) => {
    if (!message) return;
    if (!aiEnabled) {
      append("bot", "AI disabled. Enable AI to use chat.");
      return;
    }
    let outgoing = message;
    if (chatAgent !== "auto") {
      const label = {
        jobs: "1 (Jobs)",
        coding: "2 (Coding)",
        projects: "3 (Projects)",
        networking: "4 (Networking)",
      }[chatAgent];
      outgoing = `Option ${label}: ${message}`;
    }
    setBusy(true);
    try {
      if (!apiBase) throw new Error("API base not set");
      const resp = await fetch(`${apiBase}/chat`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: outgoing, session_id: sessionId }),
      });
      const data = await resp.json();
      if (!resp.ok) throw new Error(data.error || "Request failed");
      if (data.session_id) setSessionId(data.session_id);
      append("me", outgoing);
      if (data.replies?.length) {
        data.replies.forEach((t) => append("bot", t));
      } else {
        append("bot", "No reply received.");
      }
    } catch (err) {
      setErrors((prev) => [...prev, err.message]);
      append("bot", `Error: ${err.message}`);
    } finally {
      setBusy(false);
    }
  };

  const fetchSnapshot = async () => {
    setLoadingSnapshot(true);
    try {
      if (!apiBase) return;
      const resp = await fetch(`${apiBase}/data`);
      const data = await resp.json();
      if (!resp.ok) throw new Error(data.error || "Failed to fetch data");
      setSnapshot(data);
    } catch (err) {
      setErrors((prev) => [...prev, err.message]);
    } finally {
      setLoadingSnapshot(false);
    }
  };

  useEffect(() => {
    if (!apiBase) return;
    fetchSnapshot();
  }, [apiBase]);

  const handleChatSubmit = (e) => {
    e.preventDefault();
    const msg = chatInput.trim();
    setChatInput("");
    sendMessage(msg);
  };

  const handleSearch = (e) => {
    e.preventDefault();
    const query = searchText.trim();
    setSearchQuery(query);
    if (query) {
      setActiveSection("search");
    }
  };

  const handleClearSearch = () => {
    setSearchText("");
    setSearchQuery("");
  };

  const handleSaveJob = async (payload) => {
    await postJSON("/jobs", payload);
    await fetchSnapshot();
  };

  const handleSaveCoding = async (payload) => {
    await postJSON("/coding", payload);
    await fetchSnapshot();
  };

  const handleSaveProject = async (payload) => {
    await postJSON("/projects", payload);
    await fetchSnapshot();
  };

  const handleSaveNetworking = async (payload) => {
    await postJSON("/networking", payload);
    await fetchSnapshot();
  };

  const handleUpdateJobStatus = async (payload) => {
    await postJSON("/jobs/status", payload, "PATCH");
    await fetchSnapshot();
  };

  const handleUpdateGoal = async (payload) => {
    await postJSON("/goals", payload, "PATCH");
    await fetchSnapshot();
  };

  const baseTabs = useMemo(() => {
    if (!snapshot) return [];
    return [
      {
        key: "jobs",
        label: "Jobs",
        items: snapshot.job_applications || [],
        columns: [
          { key: "job_title", label: "Title" },
          { key: "company", label: "Company" },
          { key: "status", label: "Status" },
          { key: "applied_date", label: "Applied" },
          { key: "result_date", label: "Result" },
          { key: "job_link", label: "Link" },
          { key: "notes", label: "Notes" },
        ],
      },
      {
        key: "coding",
        label: "Coding",
        items: snapshot.coding_problems || [],
        columns: [
          { key: "leetcode_number", label: "Number" },
          { key: "title", label: "Title" },
          { key: "pattern", label: "Pattern" },
          { key: "difficulty", label: "Difficulty" },
          { key: "already_solved", label: "Solved" },
          { key: "problem_link", label: "Link" },
          { key: "notes", label: "Notes" },
        ],
      },
      {
        key: "projects",
        label: "Projects",
        items: snapshot.projects || [],
        columns: [
          { key: "name", label: "Name" },
          { key: "active", label: "Active" },
          { key: "tech_stack", label: "Tech Stack" },
          { key: "repo_url", label: "Repo" },
          { key: "summary", label: "Summary" },
        ],
      },
      {
        key: "networking",
        label: "Networking",
        items: snapshot.networking_contacts || [],
        columns: [
          { key: "person_name", label: "Name" },
          { key: "company", label: "Company" },
          { key: "position", label: "Role" },
          { key: "linkedin_connected", label: "LinkedIn" },
          { key: "how_met", label: "How Met" },
          { key: "notes", label: "Notes" },
        ],
      },
      {
        key: "daily_goals",
        label: "Daily Goals",
        items: snapshot.daily_goals || [],
        goalType: "daily",
        columns: [
          { key: "description", label: "Description" },
          { key: "target_date", label: "Date" },
          { key: "completed", label: "Done" },
        ],
      },
      {
        key: "weekly_goals",
        label: "Weekly Goals",
        items: snapshot.weekly_goals || [],
        goalType: "weekly",
        columns: [
          { key: "description", label: "Description" },
          { key: "target_date", label: "Week Of" },
          { key: "completed", label: "Done" },
        ],
      },
      {
        key: "monthly_goals",
        label: "Monthly Goals",
        items: snapshot.monthly_goals || [],
        goalType: "monthly",
        columns: [
          { key: "description", label: "Description" },
          { key: "target_date", label: "Month Of" },
          { key: "completed", label: "Done" },
        ],
      },
      {
        key: "meetings",
        label: "Meetings",
        items: snapshot.meetings || [],
        columns: [
          { key: "session_name", label: "Session" },
          { key: "session_type", label: "Type" },
          { key: "session_time", label: "Time" },
          { key: "location", label: "Location" },
          { key: "organizer", label: "Organizer" },
          { key: "company", label: "Company" },
        ],
      },
    ];
  }, [snapshot]);

  useEffect(() => {
    if (baseTabs.length === 0) return;
    setPagination((prev) => {
      const next = { ...prev };
      baseTabs.forEach((tab) => {
        if (!next[tab.key]) {
          next[tab.key] = { page: 1, pageSize: 10 };
        }
      });
      return next;
    });
  }, [baseTabs]);

  const getPageState = (key) => pagination[key] || { page: 1, pageSize: 10 };
  const setPage = (key, page) => {
    setPagination((prev) => ({
      ...prev,
      [key]: { ...(prev[key] || { pageSize: 10, page: 1 }), page },
    }));
  };
  const setPageSize = (key, pageSize) => {
    setPagination((prev) => ({
      ...prev,
      [key]: { ...(prev[key] || { pageSize: 10, page: 1 }), page: 1, pageSize },
    }));
  };

  const globalResults = useMemo(() => {
    if (!searchQuery) return [];
    return baseTabs
      .map((tab) => ({
        ...tab,
        matches: filterItems(tab.items, tab.columns, searchQuery),
      }))
      .filter((tab) => tab.matches.length > 0);
  }, [baseTabs, searchQuery]);

  const tabsToRender = useMemo(() => {
    if (baseTabs.length === 0) return [];
    if (!searchQuery) return baseTabs;
    const totalMatches = globalResults.reduce((sum, tab) => sum + tab.matches.length, 0);
    return [
      {
        key: "search",
        label: "Search Results",
        count: totalMatches,
        sections: globalResults,
      },
      ...baseTabs,
    ];
  }, [baseTabs, globalResults, searchQuery]);

  useEffect(() => {
    if (tabsToRender.length === 0) return;
    const found = tabsToRender.some((tab) => tab.key === activeSection);
    if (!found) setActiveSection(tabsToRender[0].key);
  }, [tabsToRender, activeSection]);

  const tabIndex = Math.max(
    0,
    tabsToRender.findIndex((tab) => tab.key === activeSection)
  );

  const summaryStats = useMemo(() => {
    if (!snapshot) return [];
    const total = totalGoals(snapshot);
    const done = countDone([
      ...(snapshot.daily_goals || []),
      ...(snapshot.weekly_goals || []),
      ...(snapshot.monthly_goals || []),
    ]);
    return [
      { label: "Applications", value: snapshot.job_applications?.length || 0 },
      { label: "Coding Problems", value: snapshot.coding_problems?.length || 0 },
      { label: "Projects", value: snapshot.projects?.length || 0 },
      { label: "Contacts", value: snapshot.networking_contacts?.length || 0 },
      { label: "Goals Done", value: `${done} / ${total}` },
      { label: "Meetings", value: snapshot.meetings?.length || 0 },
    ];
  }, [snapshot]);

  return (
    <Box minH="100vh" bgGradient="linear(to-br, teal.50, orange.50, gray.50)" px={{ base: 4, md: 8 }} py={6}>
      <Stack spacing={6}>
        <Flex
          bg="white"
          borderRadius="lg"
          boxShadow="sm"
          p={4}
          align="center"
          justify="space-between"
          wrap="wrap"
          gap={4}
        >
          <HStack spacing={4}>
            <Box
              bg="teal.500"
              color="white"
              borderRadius="md"
              px={3}
              py={2}
              fontWeight="700"
            >
              CK
            </Box>
            <Box>
              <Heading size="md">CareerKoala</Heading>
              <Text fontSize="sm" color="gray.600">
                Single-user coach for jobs, coding, projects, and networking
              </Text>
            </Box>
          </HStack>
          <VStack align="flex-end" spacing={1}>
            <HStack spacing={2}>
              <Badge colorScheme={aiEnabled ? "green" : "gray"}>
                {aiEnabled ? "AI ON" : "AI OFF"}
              </Badge>
              <Badge colorScheme="blue">API</Badge>
              <Text fontSize="xs" color="gray.500">
                {apiBase || "API base not set"}
              </Text>
            </HStack>
            <Text fontSize="xs" color="gray.500">
              Session: {sessionId || "new"}
            </Text>
          </VStack>
        </Flex>

        {!aiEnabled && (
          <Alert status="info" variant="left-accent" bg="blue.50" borderRadius="lg">
            <AlertIcon />
            AI is disabled. You can still add data; chat is hidden.
          </Alert>
        )}

        <Card borderRadius="xl" boxShadow="sm">
          <CardHeader>
            <Heading size="sm">Summary</Heading>
            <Text fontSize="sm" color="gray.600">
              High-level snapshot from your database.
            </Text>
          </CardHeader>
          <CardBody>
            {summaryStats.length === 0 ? (
              <Text color="gray.500">No data loaded yet.</Text>
            ) : (
              <SimpleGrid columns={{ base: 2, md: 3 }} spacing={4}>
                {summaryStats.map((stat) => (
                  <Stat key={stat.label} p={3} borderRadius="lg" bg="gray.50" border="1px solid" borderColor="gray.200">
                    <StatLabel color="gray.500" fontSize="xs">{stat.label}</StatLabel>
                    <StatNumber fontSize="lg">{stat.value}</StatNumber>
                  </Stat>
                ))}
              </SimpleGrid>
            )}
          </CardBody>
        </Card>

        <Card borderRadius="xl" boxShadow="sm">
          <CardHeader>
            <Heading size="sm">Quick Capture</Heading>
            <Text fontSize="sm" color="gray.600">
              Log updates fast, then review them in Detailed Records.
            </Text>
          </CardHeader>
          <CardBody>
            <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4} alignItems="start">
              <EntryCard title="Job Applications" description="Track applications and statuses.">
                <JobForm onSubmit={handleSaveJob} />
              </EntryCard>
              <EntryCard title="Coding" description="Log practice problems and patterns.">
                <CodingForm onSubmit={handleSaveCoding} />
              </EntryCard>
              <EntryCard title="Projects" description="Capture project progress and tech stack.">
                <ProjectForm onSubmit={handleSaveProject} />
              </EntryCard>
              <EntryCard title="Networking" description="Track contacts and follow-ups.">
                <NetworkingForm onSubmit={handleSaveNetworking} />
              </EntryCard>
            </SimpleGrid>
          </CardBody>
        </Card>

        <Card borderRadius="xl" boxShadow="sm">
          <CardHeader>
            <Flex align="center" justify="space-between" wrap="wrap" gap={3}>
              <Box>
                <Heading size="sm">Detailed Records</Heading>
                <Text fontSize="sm" color="gray.600">
                  Browse full records by section and search within them.
                </Text>
              </Box>
              <Button onClick={fetchSnapshot} isLoading={loadingSnapshot} colorScheme="teal">
                Refresh
              </Button>
            </Flex>
          </CardHeader>
          <CardBody>
            <Stack spacing={4}>
              <HStack spacing={2} wrap="wrap">
                <Input
                  type="search"
                  placeholder="Search across all sections"
                  value={searchText}
                  onChange={(e) => setSearchText(e.target.value)}
                  isDisabled={!snapshot}
                />
                <Button onClick={handleSearch} isDisabled={!snapshot} colorScheme="blue">
                  Search
                </Button>
                <Button variant="outline" onClick={handleClearSearch} isDisabled={!snapshot}>
                  Clear
                </Button>
              </HStack>

              {tabsToRender.length === 0 ? (
                <Text color="gray.500">No data loaded yet.</Text>
              ) : (
                <Tabs
                  index={tabIndex}
                  onChange={(idx) => setActiveSection(tabsToRender[idx].key)}
                  variant="soft-rounded"
                  colorScheme="teal"
                  isLazy
                >
                  <TabList flexWrap="wrap" gap={2}>
                    {tabsToRender.map((tab) => (
                      <Tab key={tab.key}>
                        <HStack spacing={2}>
                          <Text>{tab.label}</Text>
                          <Badge colorScheme="gray">
                            {"count" in tab ? tab.count : tab.items.length}
                          </Badge>
                        </HStack>
                      </Tab>
                    ))}
                  </TabList>
                  <TabPanels pt={4}>
                  {tabsToRender.map((tab) => {
                      if (tab.key === "search") {
                        const totalMatches = tab.count || 0;
                        return (
                          <TabPanel key={tab.key} p={0}>
                            {searchQuery.trim().length === 0 && (
                              <Text color="gray.500">Enter a keyword to search.</Text>
                            )}
                            {searchQuery.trim().length > 0 && totalMatches === 0 && (
                              <Alert status="warning" borderRadius="lg">
                                <AlertIcon />
                                No records found for "{searchQuery}". Try a different keyword.
                              </Alert>
                            )}
                            {searchQuery.trim().length > 0 && totalMatches > 0 && (
                              <Stack spacing={5}>
                                {tab.sections.map((section) => (
                                  <Box key={section.key}>
                                    <HStack mb={2} spacing={2}>
                                      <Heading size="xs">{section.label}</Heading>
                                      <Badge colorScheme="teal">{section.matches.length}</Badge>
                                    </HStack>
                                    {section.goalType ? (
                                      <GoalList
                                        items={section.matches}
                                        goalType={section.goalType}
                                        highlightTerm={searchQuery}
                                        page={getPageState(section.key).page}
                                        pageSize={getPageState(section.key).pageSize}
                                        onPageChange={(page) => setPage(section.key, page)}
                                        onPageSizeChange={(size) => setPageSize(section.key, size)}
                                        onUpdate={handleUpdateGoal}
                                      />
                                    ) : section.key === "jobs" ? (
                                      <JobList
                                        items={section.matches}
                                        highlightTerm={searchQuery}
                                        page={getPageState(section.key).page}
                                        pageSize={getPageState(section.key).pageSize}
                                        onPageChange={(page) => setPage(section.key, page)}
                                        onPageSizeChange={(size) => setPageSize(section.key, size)}
                                        onUpdate={handleUpdateJobStatus}
                                      />
                                    ) : (
                                      <RecordList
                                        items={section.matches}
                                        columns={section.columns}
                                        highlightTerm={searchQuery}
                                        page={getPageState(section.key).page}
                                        pageSize={getPageState(section.key).pageSize}
                                        onPageChange={(page) => setPage(section.key, page)}
                                        onPageSizeChange={(size) => setPageSize(section.key, size)}
                                      />
                                    )}
                                  </Box>
                                ))}
                              </Stack>
                            )}
                          </TabPanel>
                        );
                      }
                      const filtered = filterItems(tab.items, tab.columns, searchQuery);
                      const showNoMatch =
                        searchQuery.trim().length > 0 &&
                        tab.items.length > 0 &&
                        filtered.length === 0;
                      return (
                        <TabPanel key={tab.key} p={0}>
                          {showNoMatch && (
                            <Alert status="warning" mb={4} borderRadius="lg">
                              <AlertIcon />
                              No records found for "{searchQuery}". Try a different keyword.
                            </Alert>
                          )}
                          {tab.goalType ? (
                            <GoalList
                              items={filtered}
                              goalType={tab.goalType}
                              highlightTerm={searchQuery}
                              page={getPageState(tab.key).page}
                              pageSize={getPageState(tab.key).pageSize}
                              onPageChange={(page) => setPage(tab.key, page)}
                              onPageSizeChange={(size) => setPageSize(tab.key, size)}
                              onUpdate={handleUpdateGoal}
                            />
                          ) : tab.key === "jobs" ? (
                            <JobList
                              items={filtered}
                              highlightTerm={searchQuery}
                              page={getPageState(tab.key).page}
                              pageSize={getPageState(tab.key).pageSize}
                              onPageChange={(page) => setPage(tab.key, page)}
                              onPageSizeChange={(size) => setPageSize(tab.key, size)}
                              onUpdate={handleUpdateJobStatus}
                            />
                          ) : (
                            <RecordList
                              items={filtered}
                              columns={tab.columns}
                              highlightTerm={searchQuery}
                              page={getPageState(tab.key).page}
                              pageSize={getPageState(tab.key).pageSize}
                              onPageChange={(page) => setPage(tab.key, page)}
                              onPageSizeChange={(size) => setPageSize(tab.key, size)}
                            />
                          )}
                        </TabPanel>
                      );
                    })}
                  </TabPanels>
                </Tabs>
              )}
            </Stack>
          </CardBody>
        </Card>

        {aiEnabled && (
          <Card borderRadius="xl" boxShadow="sm">
            <CardHeader>
              <Heading size="sm">Chat</Heading>
              <Text fontSize="sm" color="gray.600">
                Ask the root agent or pick a specialist.
              </Text>
            </CardHeader>
            <CardBody>
              <Stack spacing={3}>
                <VStack
                  align="stretch"
                  spacing={3}
                  maxH="280px"
                  overflowY="auto"
                  border="1px solid"
                  borderColor="gray.200"
                  borderRadius="md"
                  p={3}
                  bg="gray.50"
                >
                  {chatLog.length === 0 && (
                    <Text fontSize="sm" color="gray.500">
                      No messages yet.
                    </Text>
                  )}
                  {chatLog.map((msg, idx) => (
                    <Box
                      key={idx}
                      alignSelf={msg.role === "me" ? "flex-end" : "flex-start"}
                      bg={msg.role === "me" ? "blue.50" : "green.50"}
                      border="1px solid"
                      borderColor={msg.role === "me" ? "blue.100" : "green.100"}
                      borderRadius="md"
                      px={3}
                      py={2}
                      maxW="80%"
                    >
                      <Text fontSize="sm">{msg.text}</Text>
                    </Box>
                  ))}
                </VStack>

                <Divider />

                <Stack as="form" onSubmit={handleChatSubmit} spacing={3}>
                  <HStack spacing={2} wrap="wrap">
                    <Select
                      value={chatAgent}
                      onChange={(e) => setChatAgent(e.target.value)}
                      maxW="220px"
                    >
                      <option value="auto">Main agent (auto)</option>
                      <option value="jobs">Jobs (1)</option>
                      <option value="coding">Coding (2)</option>
                      <option value="projects">Projects (3)</option>
                      <option value="networking">Networking (4)</option>
                    </Select>
                    <Button type="submit" colorScheme="teal" isDisabled={busy || !chatInput.trim()}>
                      Send
                    </Button>
                  </HStack>
                  <Textarea
                    value={chatInput}
                    onChange={(e) => setChatInput(e.target.value)}
                    placeholder="Chat with Koala"
                    minH="90px"
                  />
                </Stack>
              </Stack>
            </CardBody>
          </Card>
        )}

        {errors.length > 0 && (
          <Stack spacing={2}>
            {errors.map((err, idx) => (
              <Alert key={idx} status="error" borderRadius="lg">
                <AlertIcon />
                {err}
              </Alert>
            ))}
          </Stack>
        )}
      </Stack>
    </Box>
  );
}

function EntryCard({ title, description, children }) {
  const { isOpen, onToggle } = useDisclosure({ defaultIsOpen: false });
  return (
    <Card variant="outline" borderRadius="lg" alignSelf="start">
      <CardHeader pb={2}>
        <Flex align="center" justify="space-between" gap={2}>
          <Box>
            <Heading size="xs">{title}</Heading>
            <Text fontSize="xs" color="gray.600">
              {description}
            </Text>
          </Box>
          <IconButton
            aria-label={isOpen ? "Collapse form" : "Expand form"}
            onClick={onToggle}
            size="sm"
            variant="ghost"
            icon={isOpen ? <ChevronUpIcon /> : <ChevronDownIcon />}
          />
        </Flex>
      </CardHeader>
      <Collapse in={isOpen} animateOpacity>
        <CardBody pt={0}>{children}</CardBody>
      </Collapse>
    </Card>
  );
}

function JobForm({ onSubmit }) {
  const [state, setState] = useState({
    jobTitle: "",
    company: "",
    jobLink: "",
    appliedDate: "",
    resultDate: "",
    status: "",
    notes: "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e) => {
    e.preventDefault();
    setError("");
    setSaving(true);
    try {
      await onSubmit({
        job_title: state.jobTitle,
        company: state.company,
        job_link: state.jobLink,
        applied_date: state.appliedDate,
        result_date: state.resultDate,
        status: state.status,
        notes: state.notes,
      });
      setState({
        jobTitle: "",
        company: "",
        jobLink: "",
        appliedDate: "",
        resultDate: "",
        status: "",
        notes: "",
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Stack as="form" onSubmit={submit} spacing={3}>
      <FormControl isRequired>
        <FormLabel>Job title</FormLabel>
        <Input
          value={state.jobTitle}
          onChange={(e) => setState({ ...state, jobTitle: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Company</FormLabel>
        <Input
          value={state.company}
          onChange={(e) => setState({ ...state, company: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Job link</FormLabel>
        <Input
          value={state.jobLink}
          onChange={(e) => setState({ ...state, jobLink: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Applied date</FormLabel>
        <Input
          type="date"
          value={state.appliedDate}
          onChange={(e) => setState({ ...state, appliedDate: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Result date</FormLabel>
        <Input
          type="date"
          value={state.resultDate}
          onChange={(e) => setState({ ...state, resultDate: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Status</FormLabel>
        <Input
          value={state.status}
          onChange={(e) => setState({ ...state, status: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Notes</FormLabel>
        <Textarea
          value={state.notes}
          onChange={(e) => setState({ ...state, notes: e.target.value })}
        />
      </FormControl>
      {error && (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          {error}
        </Alert>
      )}
      <Button type="submit" colorScheme="teal" isLoading={saving}>
        Save Job
      </Button>
    </Stack>
  );
}

function CodingForm({ onSubmit }) {
  const [state, setState] = useState({
    num: "",
    title: "",
    pattern: "",
    link: "",
    difficulty: "",
    solved: false,
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e) => {
    e.preventDefault();
    setError("");
    setSaving(true);
    try {
      await onSubmit({
        leetcode_number: state.num ? Number(state.num) : 0,
        title: state.title,
        pattern: state.pattern,
        problem_link: state.link,
        difficulty: state.difficulty,
        already_solved: state.solved,
      });
      setState({ num: "", title: "", pattern: "", link: "", difficulty: "", solved: false });
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Stack as="form" onSubmit={submit} spacing={3}>
      <FormControl>
        <FormLabel>LeetCode number</FormLabel>
        <Input
          value={state.num}
          onChange={(e) => setState({ ...state, num: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Title</FormLabel>
        <Input
          value={state.title}
          onChange={(e) => setState({ ...state, title: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Pattern</FormLabel>
        <Input
          value={state.pattern}
          onChange={(e) => setState({ ...state, pattern: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Link</FormLabel>
        <Input
          value={state.link}
          onChange={(e) => setState({ ...state, link: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Difficulty</FormLabel>
        <Input
          value={state.difficulty}
          onChange={(e) => setState({ ...state, difficulty: e.target.value })}
        />
      </FormControl>
      <Checkbox
        isChecked={state.solved}
        onChange={(e) => setState({ ...state, solved: e.target.checked })}
      >
        Already solved
      </Checkbox>
      {error && (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          {error}
        </Alert>
      )}
      <Button type="submit" colorScheme="teal" isLoading={saving}>
        Save Problem
      </Button>
    </Stack>
  );
}

function ProjectForm({ onSubmit }) {
  const [state, setState] = useState({
    name: "",
    repo: "",
    active: false,
    tech: "",
    summary: "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e) => {
    e.preventDefault();
    setError("");
    setSaving(true);
    try {
      const techStack = state.tech
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean);
      await onSubmit({
        name: state.name,
        repo_url: state.repo,
        active: state.active,
        tech_stack: techStack,
        summary: state.summary,
      });
      setState({ name: "", repo: "", active: false, tech: "", summary: "" });
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Stack as="form" onSubmit={submit} spacing={3}>
      <FormControl>
        <FormLabel>Project name</FormLabel>
        <Input
          value={state.name}
          onChange={(e) => setState({ ...state, name: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Repo URL</FormLabel>
        <Input
          value={state.repo}
          onChange={(e) => setState({ ...state, repo: e.target.value })}
        />
      </FormControl>
      <Checkbox
        isChecked={state.active}
        onChange={(e) => setState({ ...state, active: e.target.checked })}
      >
        Active
      </Checkbox>
      <FormControl>
        <FormLabel>Tech stack (comma separated)</FormLabel>
        <Input
          value={state.tech}
          onChange={(e) => setState({ ...state, tech: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Summary</FormLabel>
        <Textarea
          value={state.summary}
          onChange={(e) => setState({ ...state, summary: e.target.value })}
        />
      </FormControl>
      {error && (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          {error}
        </Alert>
      )}
      <Button type="submit" colorScheme="teal" isLoading={saving}>
        Save Project
      </Button>
    </Stack>
  );
}

function NetworkingForm({ onSubmit }) {
  const [state, setState] = useState({
    name: "",
    howMet: "",
    connected: false,
    company: "",
    position: "",
    notes: "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const submit = async (e) => {
    e.preventDefault();
    setError("");
    setSaving(true);
    try {
      await onSubmit({
        person_name: state.name,
        how_met: state.howMet,
        linkedin_connected: state.connected,
        company: state.company,
        position: state.position,
        notes: state.notes,
      });
      setState({ name: "", howMet: "", connected: false, company: "", position: "", notes: "" });
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Stack as="form" onSubmit={submit} spacing={3}>
      <FormControl>
        <FormLabel>Name</FormLabel>
        <Input
          value={state.name}
          onChange={(e) => setState({ ...state, name: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>How you met</FormLabel>
        <Input
          value={state.howMet}
          onChange={(e) => setState({ ...state, howMet: e.target.value })}
        />
      </FormControl>
      <Checkbox
        isChecked={state.connected}
        onChange={(e) => setState({ ...state, connected: e.target.checked })}
      >
        LinkedIn connected
      </Checkbox>
      <FormControl>
        <FormLabel>Company</FormLabel>
        <Input
          value={state.company}
          onChange={(e) => setState({ ...state, company: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Position</FormLabel>
        <Input
          value={state.position}
          onChange={(e) => setState({ ...state, position: e.target.value })}
        />
      </FormControl>
      <FormControl>
        <FormLabel>Notes</FormLabel>
        <Textarea
          value={state.notes}
          onChange={(e) => setState({ ...state, notes: e.target.value })}
        />
      </FormControl>
      {error && (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          {error}
        </Alert>
      )}
      <Button type="submit" colorScheme="teal" isLoading={saving}>
        Save Contact
      </Button>
    </Stack>
  );
}

function RecordList({
  items = [],
  columns = [],
  highlightTerm = "",
  page = 1,
  pageSize = 10,
  onPageChange,
  onPageSizeChange,
}) {
  if (!items.length) {
    return <Text color="gray.500">No entries yet.</Text>;
  }
  const { visibleItems, totalPages, currentPage } = paginate(items, page, pageSize);
  return (
    <Stack spacing={3}>
      <PaginationBar
        totalItems={items.length}
        page={currentPage}
        pageSize={pageSize}
        totalPages={totalPages}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
      <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
        {visibleItems.map((row) => (
          <Card key={row.id || JSON.stringify(row)} variant="outline" borderRadius="lg">
            <CardBody>
              <Stack spacing={2}>
                {columns.map((c) => {
                  const value = formatCellValue(row[c.key]);
                  const content = value
                    ? renderHighlightedText(value, highlightTerm)
                    : "-";
                  return isLongField(c.key) ? (
                    <Box key={c.key}>
                      <Text fontSize="xs" color="gray.500" mb={1}>
                        {c.label}
                      </Text>
                      <Text fontSize="sm" fontWeight="600" whiteSpace="pre-wrap">
                        {content}
                      </Text>
                    </Box>
                  ) : (
                    <Flex key={c.key} justify="space-between" gap={6}>
                      <Text fontSize="xs" color="gray.500">
                        {c.label}
                      </Text>
                      <Text fontSize="sm" fontWeight="600" textAlign="right">
                        {content}
                      </Text>
                    </Flex>
                  );
                })}
              </Stack>
            </CardBody>
          </Card>
        ))}
      </SimpleGrid>
    </Stack>
  );
}

function JobList({
  items = [],
  highlightTerm = "",
  page = 1,
  pageSize = 10,
  onPageChange,
  onPageSizeChange,
  onUpdate,
}) {
  const [savingId, setSavingId] = useState(null);
  const [error, setError] = useState("");
  const [showRejected, setShowRejected] = useState(false);
  const filteredItems = showRejected
    ? items
    : items.filter((job) => !String(job.status || "").toLowerCase().includes("reject"));
  const { visibleItems, totalPages, currentPage } = paginate(filteredItems, page, pageSize);

  useEffect(() => {
    if (currentPage !== page && onPageChange) {
      onPageChange(currentPage);
    }
  }, [currentPage, page, onPageChange]);

  const handleReject = async (jobId) => {
    setError("");
    setSavingId(jobId);
    try {
      await onUpdate({ id: jobId, status: "rejected" });
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingId(null);
    }
  };

  if (!items.length) {
    return <Text color="gray.500">No entries yet.</Text>;
  }

  return (
    <Stack spacing={3}>
      <HStack justify="space-between" wrap="wrap">
        <HStack spacing={2}>
          <Switch
            isChecked={showRejected}
            onChange={(e) => setShowRejected(e.target.checked)}
            colorScheme="teal"
          />
          <Text fontSize="sm" color="gray.600">
            Show rejected
          </Text>
        </HStack>
        {filteredItems.length === 0 && (
          <Text fontSize="xs" color="gray.500">
            No non-rejected jobs.
          </Text>
        )}
      </HStack>
      <PaginationBar
        totalItems={filteredItems.length}
        page={currentPage}
        pageSize={pageSize}
        totalPages={totalPages}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
      {error && (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          {error}
        </Alert>
      )}
      <Box
        display={{ base: "none", md: "grid" }}
        gridTemplateColumns="2.4fr 1.4fr 0.9fr 0.9fr"
        gap={3}
        fontSize="xs"
        color="gray.500"
        textTransform="uppercase"
      >
        <Text>Role</Text>
        <Text>Company</Text>
        <Text>Status</Text>
        <Text>Actions</Text>
      </Box>
      {visibleItems.map((job, idx) => {
        const status = (job.status || "").toString();
        const statusLower = status.toLowerCase();
        const title = job.job_title || "Untitled";
        const company = job.company || "-";
        const applied = formatCellValue(job.applied_date) || "-";
        const result = formatCellValue(job.result_date) || "-";
        const notes = formatCellValue(job.notes);
        const isRejected = statusLower.includes("reject");
        const statusLabel = isRejected ? "Rejected" : status || "Applied";
        const statusDate = isRejected ? result : applied;
        return (
          <Box
            key={job.id}
            borderBottom={idx === visibleItems.length - 1 ? "none" : "1px solid"}
            borderColor="gray.200"
            py={2}
            display="grid"
            gridTemplateColumns={{ base: "1fr", md: "2.4fr 1.4fr 0.9fr 0.9fr" }}
            gap={3}
          >
            <Box>
              <Text fontSize="xs" color="gray.500" display={{ base: "block", md: "none" }}>
                Role
              </Text>
              <Text fontWeight="600" fontSize="sm" noOfLines={1}>
                {job.job_link ? (
                  <a
                    href={job.job_link}
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{ color: "#2b6cb0" }}
                  >
                    {renderHighlightedText(title, highlightTerm)}
                  </a>
                ) : (
                  renderHighlightedText(title, highlightTerm)
                )}
              </Text>
              {notes && (
                <Text fontSize="xs" color="gray.500" noOfLines={1}>
                  Notes: {renderHighlightedText(notes, highlightTerm)}
                </Text>
              )}
            </Box>
            <Box>
              <Text fontSize="xs" color="gray.500" display={{ base: "block", md: "none" }}>
                Company
              </Text>
              <Text fontSize="sm" color="gray.600" noOfLines={1}>
                {renderHighlightedText(company, highlightTerm)}
              </Text>
            </Box>
            <Box>
              <Text fontSize="xs" color="gray.500" display={{ base: "block", md: "none" }}>
                Status
              </Text>
              <Badge colorScheme={statusColor(statusLower)}>
                {renderHighlightedText(statusLabel, highlightTerm)}
              </Badge>
            </Box>
            <Box>
              <Text fontSize="xs" color="gray.500" display={{ base: "block", md: "none" }}>
                Actions
              </Text>
              <Button
                size="sm"
                colorScheme="red"
                variant="outline"
                onClick={() => handleReject(job.id)}
                isDisabled={statusLower === "rejected"}
                isLoading={savingId === job.id}
              >
                Mark rejected
              </Button>
            </Box>
          </Box>
        );
      })}
    </Stack>
  );
}

function GoalList({
  items = [],
  goalType,
  highlightTerm = "",
  page = 1,
  pageSize = 10,
  onPageChange,
  onPageSizeChange,
  onUpdate,
}) {
  const [drafts, setDrafts] = useState({});
  const [savingId, setSavingId] = useState(null);
  const [error, setError] = useState("");
  const { visibleItems, totalPages, currentPage } = paginate(items, page, pageSize);

  useEffect(() => {
    if (currentPage !== page && onPageChange) {
      onPageChange(currentPage);
    }
  }, [currentPage, page, onPageChange]);

  useEffect(() => {
    setDrafts((prev) => {
      const next = { ...prev };
      items.forEach((goal) => {
        next[goal.id] = {
          description: goal.description || "",
          completed: !!goal.completed,
        };
      });
      return next;
    });
  }, [items]);

  const handleSave = async (goalId) => {
    setError("");
    const draft = drafts[goalId];
    if (!draft) return;
    setSavingId(goalId);
    try {
      await onUpdate({
        type: goalType,
        id: goalId,
        description: draft.description,
        completed: draft.completed,
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setSavingId(null);
    }
  };

  if (!items.length) {
    return <Text color="gray.500">No entries yet.</Text>;
  }

  return (
    <Stack spacing={3}>
      <PaginationBar
        totalItems={items.length}
        page={currentPage}
        pageSize={pageSize}
        totalPages={totalPages}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
      {error && (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          {error}
        </Alert>
      )}
      {visibleItems.map((goal, idx) => {
        const draft = drafts[goal.id] || {
          description: goal.description || "",
          completed: !!goal.completed,
        };
        const highlight =
          highlightTerm &&
          draft.description.toLowerCase().includes(highlightTerm.toLowerCase());
        return (
          <Box
            key={goal.id}
            borderBottom={idx === visibleItems.length - 1 ? "none" : "1px solid"}
            borderColor="gray.200"
            pb={3}
          >
            <Stack spacing={3}>
              <HStack justify="space-between" align="center" wrap="wrap">
                <HStack spacing={2}>
                  <Badge colorScheme="purple" textTransform="capitalize">
                    {goalType}
                  </Badge>
                  <Badge colorScheme="gray">
                    {goal.target_date ? goal.target_date : "No date"}
                  </Badge>
                </HStack>
                <Checkbox
                  isChecked={draft.completed}
                  onChange={(e) =>
                    setDrafts((prev) => ({
                      ...prev,
                      [goal.id]: { ...draft, completed: e.target.checked },
                    }))
                  }
                >
                  Completed
                </Checkbox>
              </HStack>
              <FormControl>
                <FormLabel fontSize="xs" color="gray.500">
                  Description
                </FormLabel>
                <Input
                  value={draft.description}
                  onChange={(e) =>
                    setDrafts((prev) => ({
                      ...prev,
                      [goal.id]: { ...draft, description: e.target.value },
                    }))
                  }
                  bg={highlight ? "yellow.100" : "white"}
                />
              </FormControl>
              <Button
                size="sm"
                colorScheme="teal"
                alignSelf="flex-start"
                onClick={() => handleSave(goal.id)}
                isLoading={savingId === goal.id}
              >
                Update
              </Button>
            </Stack>
          </Box>
        );
      })}
    </Stack>
  );
}

function formatCellValue(val) {
  if (val === null || val === undefined) return "";
  if (Array.isArray(val)) return val.join(", ");
  if (typeof val === "boolean") return val ? "Yes" : "No";
  return String(val);
}

function statusColor(status) {
  if (status.includes("reject")) return "red";
  if (status.includes("offer")) return "green";
  if (status.includes("interview")) return "orange";
  if (status.includes("applied")) return "blue";
  return "gray";
}

function isLongField(key) {
  return [
    "summary",
    "notes",
    "job_link",
    "problem_link",
    "repo_url",
    "location",
  ].includes(key);
}

function filterItems(items = [], columns = [], query) {
  const term = query.trim().toLowerCase();
  if (!term) return items;
  return items.filter((row) =>
    columns.some((c) => {
      const val = formatCellValue(row[c.key]);
      return String(val).toLowerCase().includes(term);
    })
  );
}

function renderHighlightedText(text, term) {
  const trimmed = term.trim();
  if (!trimmed) return text;
  const safeTerm = escapeRegExp(trimmed);
  const regex = new RegExp(`(${safeTerm})`, "ig");
  const parts = text.split(regex);
  const lowerTerm = trimmed.toLowerCase();
  return parts.map((part, idx) => {
    if (part.toLowerCase() === lowerTerm) {
      return (
        <Box
          as="mark"
          key={`${part}-${idx}`}
          bg="yellow.200"
          color="black"
          borderRadius="sm"
          px="1"
        >
          {part}
        </Box>
      );
    }
    return <span key={`${part}-${idx}`}>{part}</span>;
  });
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function paginate(items, page, pageSize) {
  const safeSize = Number.isFinite(pageSize) && pageSize > 0 ? pageSize : items.length;
  const totalPages = Math.max(1, Math.ceil(items.length / safeSize));
  const safePage = Math.min(Math.max(page, 1), totalPages);
  const start = (safePage - 1) * safeSize;
  const visibleItems = items.slice(start, start + safeSize);
  return { visibleItems, totalPages, currentPage: safePage };
}

function PaginationBar({
  totalItems,
  page,
  pageSize,
  totalPages,
  onPageChange,
  onPageSizeChange,
}) {
  if (!totalItems) return null;
  const pages = buildPagination(totalPages, page);
  const handlePageChange = (nextPage) => {
    if (!onPageChange) return;
    const clamped = Math.min(Math.max(nextPage, 1), totalPages);
    onPageChange(clamped);
  };
  return (
    <HStack justify="space-between" wrap="wrap">
      <HStack spacing={1} wrap="wrap">
        <Button size="sm" variant="ghost" onClick={() => handlePageChange(1)} isDisabled={page <= 1}>
          {"<<"}
        </Button>
        <Button size="sm" variant="ghost" onClick={() => handlePageChange(page - 1)} isDisabled={page <= 1}>
          {"<"}
        </Button>
        {pages.map((item, idx) =>
          item === "ellipsis" ? (
            <Text key={`ellipsis-${idx}`} color="gray.500" px={2}>
              ...
            </Text>
          ) : (
            <Button
              key={`page-${item}`}
              size="sm"
              variant={item === page ? "solid" : "ghost"}
              colorScheme={item === page ? "teal" : "gray"}
              onClick={() => handlePageChange(item)}
            >
              {item}
            </Button>
          )
        )}
        <Button
          size="sm"
          variant="ghost"
          onClick={() => handlePageChange(page + 1)}
          isDisabled={page >= totalPages}
        >
          {">"}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => handlePageChange(totalPages)}
          isDisabled={page >= totalPages}
        >
          {">>"}
        </Button>
      </HStack>
      <HStack spacing={2}>
        <Text fontSize="sm" color="gray.600">
          {page} / {totalPages}
        </Text>
        <Select
          size="sm"
          value={pageSize}
          onChange={(e) => onPageSizeChange?.(Number(e.target.value))}
          maxW="130px"
        >
          <option value={10}>10 per page</option>
          <option value={25}>25 per page</option>
          <option value={50}>50 per page</option>
          <option value={100}>100 per page</option>
        </Select>
      </HStack>
    </HStack>
  );
}

function buildPagination(totalPages, currentPage) {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }
  const pages = [1];
  const left = Math.max(currentPage - 1, 2);
  const right = Math.min(currentPage + 1, totalPages - 1);
  if (left > 2) {
    pages.push("ellipsis");
  }
  for (let i = left; i <= right; i += 1) {
    pages.push(i);
  }
  if (right < totalPages - 1) {
    pages.push("ellipsis");
  }
  pages.push(totalPages);
  return pages;
}

function countDone(goals) {
  return goals.filter((g) => g.completed).length;
}

function totalGoals(snapshot) {
  return (
    (snapshot.daily_goals?.length || 0) +
    (snapshot.weekly_goals?.length || 0) +
    (snapshot.monthly_goals?.length || 0)
  );
}
