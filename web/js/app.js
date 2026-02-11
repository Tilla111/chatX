
const API_BASE = "/api/v1";

const state = {
  currentUserId: null,
  chats: [],
  users: [],
  groupMemberPool: [],
  groupSelectedMemberIds: new Set(),
  selectedChatId: null,
  messages: [],
  members: [],
  memberCandidates: [],
  ws: null,
  wsTimer: null,
  manualWsClose: false,
  wsConnected: false,
  health: null,
};

const els = {
  healthBadge: document.getElementById("healthBadge"),
  wsBadge: document.getElementById("wsBadge"),
  userIdInput: document.getElementById("userIdInput"),
  connectBtn: document.getElementById("connectBtn"),
  disconnectBtn: document.getElementById("disconnectBtn"),
  refreshAllBtn: document.getElementById("refreshAllBtn"),

  chatSearchInput: document.getElementById("chatSearchInput"),
  userSearchInput: document.getElementById("userSearchInput"),
  chatList: document.getElementById("chatList"),
  userList: document.getElementById("userList"),
  chatCountBadge: document.getElementById("chatCountBadge"),

  newPrivateBtn: document.getElementById("newPrivateBtn"),
  newGroupBtn: document.getElementById("newGroupBtn"),
  membersBtn: document.getElementById("membersBtn"),
  editGroupBtn: document.getElementById("editGroupBtn"),
  markReadBtn: document.getElementById("markReadBtn"),
  deleteChatBtn: document.getElementById("deleteChatBtn"),

  activeChatName: document.getElementById("activeChatName"),
  activeChatInfo: document.getElementById("activeChatInfo"),

  messagesList: document.getElementById("messagesList"),
  composerForm: document.getElementById("composerForm"),
  messageInput: document.getElementById("messageInput"),
  sendMessageBtn: document.getElementById("sendMessageBtn"),

  privateModal: document.getElementById("privateModal"),
  privateForm: document.getElementById("privateForm"),
  privateReceiverSelect: document.getElementById("privateReceiverSelect"),

  groupModal: document.getElementById("groupModal"),
  groupForm: document.getElementById("groupForm"),
  groupNameInput: document.getElementById("groupNameInput"),
  groupDescInput: document.getElementById("groupDescInput"),
  groupMemberSearchInput: document.getElementById("groupMemberSearchInput"),
  groupMemberChecklist: document.getElementById("groupMemberChecklist"),
  groupMemberCount: document.getElementById("groupMemberCount"),

  editGroupModal: document.getElementById("editGroupModal"),
  editGroupForm: document.getElementById("editGroupForm"),
  editGroupNameInput: document.getElementById("editGroupNameInput"),
  editGroupDescInput: document.getElementById("editGroupDescInput"),

  membersModal: document.getElementById("membersModal"),
  memberSearchInput: document.getElementById("memberSearchInput"),
  memberSearchList: document.getElementById("memberSearchList"),
  membersList: document.getElementById("membersList"),

  toastStack: document.getElementById("toastStack"),
};

document.addEventListener("DOMContentLoaded", init);

function init() {
  bindEvents();
  renderHealthBadge();
  renderWsBadge();
  renderEmptyLists();
  renderChatMeta();
  renderComposerState();

  const savedUserId = Number(localStorage.getItem("chatx_user_id"));
  if (Number.isInteger(savedUserId) && savedUserId > 0) {
    els.userIdInput.value = String(savedUserId);
    connectSession();
  } else {
    fetchHealth();
  }

  setInterval(() => {
    fetchHealth();
  }, 30000);
}

function bindEvents() {
  els.connectBtn.addEventListener("click", connectSession);
  els.disconnectBtn.addEventListener("click", disconnectSession);
  els.refreshAllBtn.addEventListener("click", refreshAllData);

  els.chatSearchInput.addEventListener("input", debounce(() => {
    refreshChats();
  }, 260));
  els.userSearchInput.addEventListener("input", debounce(() => {
    refreshUsers();
  }, 260));
  els.groupMemberSearchInput.addEventListener("input", debounce(() => {
    loadGroupMemberPool(els.groupMemberSearchInput.value.trim()).catch((error) => {
      toast(error.message, "error");
    });
  }, 260));
  els.memberSearchInput.addEventListener("input", debounce(() => {
    loadMemberCandidates(els.memberSearchInput.value.trim()).catch((error) => {
      toast(error.message, "error");
    });
  }, 260));

  els.newGroupBtn.addEventListener("click", () => openGroupModal());
  els.membersBtn.addEventListener("click", () => openMembersModal());
  els.editGroupBtn.addEventListener("click", () => openEditGroupModal());
  els.markReadBtn.addEventListener("click", async () => {
    await markCurrentChatAsRead(false);
  });
  els.deleteChatBtn.addEventListener("click", deleteCurrentChat);

  els.privateForm.addEventListener("submit", submitPrivateChat);
  els.groupForm.addEventListener("submit", submitGroupChat);
  els.editGroupForm.addEventListener("submit", submitGroupUpdate);
  els.composerForm.addEventListener("submit", submitMessage);

  els.chatList.addEventListener("click", async (event) => {
    const item = event.target.closest("[data-chat-id]");
    if (!item) return;
    const chatId = Number(item.dataset.chatId);
    if (!chatId) return;
    await selectChat(chatId);
  });

  els.userList.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-action='start-private']");
    if (!button) return;
    const receiverId = Number(button.dataset.userId);
    if (!receiverId) return;
    await createPrivateChat(receiverId);
  });

  els.messagesList.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-action]");
    if (!button) return;
    const messageId = Number(button.dataset.messageId);
    if (!messageId) return;

    if (button.dataset.action === "edit-message") {
      await editMessage(messageId);
    }
    if (button.dataset.action === "delete-message") {
      await deleteMessage(messageId);
    }
  });

  els.membersList.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-action='remove-member']");
    if (!button) return;
    const userId = Number(button.dataset.userId);
    if (!userId) return;
    await removeMember(userId);
  });
  els.memberSearchList.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-action='add-member']");
    if (!button) return;
    const userId = Number(button.dataset.userId);
    if (!userId) return;
    await addMember(userId);
  });

  document.querySelectorAll("[data-close-modal]").forEach((button) => {
    button.addEventListener("click", () => {
      closeModal(button.dataset.closeModal);
    });
  });

  document.querySelectorAll(".modal").forEach((modal) => {
    modal.addEventListener("click", (event) => {
      if (event.target === modal) {
        modal.classList.add("hidden");
      }
    });
  });
}

async function connectSession() {
  const userId = Number(els.userIdInput.value.trim());
  if (!Number.isInteger(userId) || userId <= 0) {
    toast("X-User-ID musbat son bo'lishi kerak.", "error");
    return;
  }

  state.currentUserId = userId;
  localStorage.setItem("chatx_user_id", String(userId));
  toast(`User #${userId} bilan ulandingiz.`, "ok");

  await refreshAllData();
  connectWebSocket();
}

function disconnectSession() {
  state.currentUserId = null;
  state.chats = [];
  state.users = [];
  state.groupMemberPool = [];
  state.groupSelectedMemberIds = new Set();
  state.messages = [];
  state.members = [];
  state.memberCandidates = [];
  state.selectedChatId = null;

  if (state.wsTimer) {
    clearTimeout(state.wsTimer);
    state.wsTimer = null;
  }

  state.manualWsClose = true;
  if (state.ws) {
    state.ws.close();
    state.ws = null;
  }
  state.wsConnected = false;

  localStorage.removeItem("chatx_user_id");
  renderWsBadge();
  renderChatList();
  renderUserList();
  renderMessages();
  renderChatMeta();
  renderComposerState();
  toast("Sessiya uzildi.", "info");
}

async function refreshAllData() {
  if (!state.currentUserId) {
    await fetchHealth();
    return;
  }

  const [healthRes, usersRes, chatsRes] = await Promise.allSettled([
    fetchHealth(),
    refreshUsers(),
    refreshChats(),
  ]);

  if (healthRes.status === "rejected") {
    toast(healthRes.reason.message, "error");
  }
  if (usersRes.status === "rejected") {
    toast(usersRes.reason.message, "error");
  }
  if (chatsRes.status === "rejected") {
    toast(chatsRes.reason.message, "error");
  }

  if (state.selectedChatId) {
    const exists = state.chats.some((chat) => chat.chatId === state.selectedChatId);
    if (!exists) {
      state.selectedChatId = null;
      state.messages = [];
      renderMessages();
      renderChatMeta();
      renderComposerState();
    }
  }
}

async function fetchHealth() {
  const health = await apiRequest("/health", { auth: false });
  state.health = health;
  renderHealthBadge();
}

async function refreshUsers() {
  if (!state.currentUserId) return;

  const query = new URLSearchParams();
  query.set("limit", "20");
  query.set("offset", "0");

  const search = els.userSearchInput.value.trim();
  if (search) {
    query.set("search", search);
  }

  const data = await apiRequest(`/users?${query.toString()}`);
  state.users = asArray(data).map(normalizeUser).filter((user) => user.id > 0);
  renderUserList();
}

async function refreshChats() {
  if (!state.currentUserId) return;

  const query = new URLSearchParams();
  const search = els.chatSearchInput.value.trim();
  if (search) {
    query.set("search", search);
  }

  const path = query.toString() ? `/chats?${query.toString()}` : "/chats";
  const data = await apiRequest(path);
  state.chats = asArray(data).map(normalizeChat).filter((chat) => chat.chatId > 0);
  renderChatList();
}

async function selectChat(chatId) {
  if (!state.currentUserId) {
    toast("Avval X-User-ID bilan ulaning.", "error");
    return;
  }

  state.selectedChatId = chatId;
  renderChatList();
  renderChatMeta();
  renderComposerState();
  setMessagesLoading();

  try {
    const rawMessages = await apiRequest(`/chats/${chatId}/messages`);
    state.messages = asArray(rawMessages).map((message) => normalizeMessage(message, chatId));
    renderMessages();
    markCurrentChatAsRead(true).catch(() => {
      toast("Read holatini yangilab bo'lmadi.", "error");
    });
  } catch (error) {
    state.messages = [];
    renderMessages();
    toast(error.message, "error");
  }
}
async function markCurrentChatAsRead(silent) {
  if (!state.selectedChatId || !state.currentUserId) return;
  await apiRequest(`/messages/chats/${state.selectedChatId}/read`, { method: "PATCH" });

  const chat = getSelectedChat();
  if (chat) {
    chat.unreadCount = 0;
  }
  renderChatList();
  if (!silent) {
    toast("Chat read holatiga o'tkazildi.", "ok");
  }
}

async function submitMessage(event) {
  event.preventDefault();
  if (!state.currentUserId) {
    toast("Avval X-User-ID bilan ulaning.", "error");
    return;
  }
  if (!state.selectedChatId) {
    toast("Chat tanlanmagan.", "error");
    return;
  }

  const messageText = els.messageInput.value.trim();
  if (!messageText) return;

  try {
    const created = await apiRequest("/messages", {
      method: "POST",
      body: {
        chat_id: state.selectedChatId,
        message_text: messageText,
      },
    });

    const normalized = normalizeMessage(created, state.selectedChatId);
    state.messages.push(normalized);
    els.messageInput.value = "";
    renderMessages(true);
    await refreshChats();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function editMessage(messageId) {
  const message = state.messages.find((item) => item.id === messageId);
  if (!message) return;
  if (message.senderId !== state.currentUserId) {
    toast("Faqat o'zingiz yuborgan xabarni tahrirlaysiz.", "error");
    return;
  }

  const nextText = prompt("Yangi xabar matni:", message.content);
  if (nextText === null) return;
  const trimmed = nextText.trim();
  if (!trimmed || trimmed === message.content) return;

  try {
    await apiRequest(`/messages/${messageId}`, {
      method: "PATCH",
      body: { message_text: trimmed },
    });
    message.content = trimmed;
    renderMessages();
    toast("Xabar yangilandi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function deleteMessage(messageId) {
  const message = state.messages.find((item) => item.id === messageId);
  if (!message) return;

  const ok = confirm("Xabarni o'chirishni tasdiqlaysizmi?");
  if (!ok) return;

  try {
    await apiRequest(`/messages/${messageId}`, { method: "DELETE" });
    state.messages = state.messages.filter((item) => item.id !== messageId);
    renderMessages();
    await refreshChats();
    toast("Xabar o'chirildi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function createPrivateChat(receiverId) {
  if (!state.currentUserId) {
    toast("Avval X-User-ID bilan ulaning.", "error");
    return;
  }

  try {
    const response = await apiRequest("/chats", {
      method: "POST",
      body: { receiver_id: receiverId },
    });
    const chatId = Number(response?.chat_id || response?.chatId || response);
    await refreshChats();
    if (chatId > 0) {
      await selectChat(chatId);
    }
    closeModal("privateModal");
    toast("Private chat yaratildi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function submitPrivateChat(event) {
  event.preventDefault();
  const receiverId = Number(els.privateReceiverSelect.value);
  if (!receiverId) {
    toast("Receiver tanlang.", "error");
    return;
  }
  await createPrivateChat(receiverId);
}

function openPrivateModal() {
  if (!ensureSession()) return;
  if (!state.users.length) {
    toast("Avval users ro'yxatini yangilang.", "info");
  }

  els.privateReceiverSelect.innerHTML = "";
  const users = state.users.filter((user) => user.id !== state.currentUserId);
  if (!users.length) {
    const option = document.createElement("option");
    option.value = "";
    option.textContent = "Foydalanuvchi topilmadi";
    els.privateReceiverSelect.appendChild(option);
  } else {
    users.forEach((user) => {
      const option = document.createElement("option");
      option.value = String(user.id);
      option.textContent = `${user.username} (#${user.id})`;
      els.privateReceiverSelect.appendChild(option);
    });
  }

  openModal("privateModal");
}

async function openGroupModal() {
  if (!ensureSession()) return;

  els.groupNameInput.value = "";
  els.groupDescInput.value = "";
  els.groupMemberSearchInput.value = "";
  state.groupSelectedMemberIds = new Set();
  state.groupMemberPool = [];
  els.groupMemberChecklist.innerHTML = `<div class="empty">A'zolar yuklanmoqda...</div>`;
  updateGroupMemberCount();
  openModal("groupModal");
  try {
    await loadGroupMemberPool("");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function submitGroupChat(event) {
  event.preventDefault();

  const name = els.groupNameInput.value.trim();
  const description = els.groupDescInput.value.trim();
  const memberIds = getSelectedGroupMembers();

  if (!name) {
    toast("Group nomi bo'sh bo'lmasin.", "error");
    return;
  }
  if (memberIds.length === 0) {
    toast("Kamida bitta a'zo tanlang.", "error");
    return;
  }

  try {
    const response = await apiRequest("/groups", {
      method: "POST",
      body: {
        name,
        description,
        member_ids: memberIds,
      },
    });
    const chatId = Number(response?.chat_id || response?.chatId || response);
    await refreshChats();
    if (chatId > 0) {
      await selectChat(chatId);
    }
    closeModal("groupModal");
    toast("Group yaratildi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

function openEditGroupModal() {
  const chat = getSelectedChat();
  if (!chat) {
    toast("Avval chat tanlang.", "error");
    return;
  }
  if (chat.chatType !== "group") {
    toast("Bu amal faqat group chat uchun.", "error");
    return;
  }

  els.editGroupNameInput.value = chat.chatName;
  els.editGroupDescInput.value = "";
  openModal("editGroupModal");
}

async function submitGroupUpdate(event) {
  event.preventDefault();
  const chat = getSelectedChat();
  if (!chat) return;

  const name = els.editGroupNameInput.value.trim();
  const description = els.editGroupDescInput.value.trim();
  if (!name) {
    toast("Group nomi bo'sh bo'lmasin.", "error");
    return;
  }

  try {
    await apiRequest(`/groups/${chat.chatId}`, {
      method: "PATCH",
      body: { name, description },
    });
    await refreshChats();
    const updated = state.chats.find((item) => item.chatId === chat.chatId);
    if (updated) {
      state.selectedChatId = updated.chatId;
    }
    renderChatMeta();
    closeModal("editGroupModal");
    toast("Group yangilandi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function deleteCurrentChat() {
  const chat = getSelectedChat();
  if (!chat) {
    toast("Avval chat tanlang.", "error");
    return;
  }

  const ok = confirm(`"${chat.chatName}" chatini o'chirishni tasdiqlaysizmi?`);
  if (!ok) return;

  try {
    await apiRequest(`/chats/${chat.chatId}`, { method: "DELETE" });
    state.selectedChatId = null;
    state.messages = [];
    await refreshChats();
    renderMessages();
    renderChatMeta();
    renderComposerState();
    toast("Chat o'chirildi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function openMembersModal() {
  const chat = getSelectedChat();
  if (!chat) {
    toast("Avval chat tanlang.", "error");
    return;
  }
  if (chat.chatType !== "group") {
    toast("A'zolar ro'yxati faqat group chatda mavjud.", "error");
    return;
  }

  state.memberCandidates = [];
  els.memberSearchInput.value = "";
  renderMemberCandidateList();

  try {
    const response = await apiRequest(`/groups/${chat.chatId}/members`);
    state.members = asArray(response).map(normalizeUser).filter((user) => user.id > 0);
    renderMembersList();
    openModal("membersModal");
    await loadMemberCandidates("");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function addMember(userId) {
  const chat = getSelectedChat();
  if (!chat) return;
  if (chat.chatType !== "group") {
    toast("A'zo qo'shish faqat group chat uchun.", "error");
    return;
  }

  try {
    await apiRequest(`/groups/${chat.chatId}/members`, {
      method: "POST",
      body: { user_id: userId },
    });
    toast("A'zo groupga qo'shildi.", "ok");
    await openMembersModal();
    await refreshChats();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function removeMember(userId) {
  const chat = getSelectedChat();
  if (!chat) return;
  const member = state.members.find((item) => item.id === userId);
  const title = member ? member.username : `#${userId}`;

  const ok = confirm(`${title} ni groupdan chiqarishni tasdiqlaysizmi?`);
  if (!ok) return;

  try {
    await apiRequest(`/groups/${chat.chatId}/${userId}/member`, { method: "DELETE" });
    toast("A'zo chiqarildi.", "ok");
    await openMembersModal();
    await refreshChats();

    if (userId === state.currentUserId) {
      closeModal("membersModal");
      state.selectedChatId = null;
      state.messages = [];
      renderMessages();
      renderChatMeta();
      renderComposerState();
    }
  } catch (error) {
    toast(error.message, "error");
  }
}
function connectWebSocket() {
  if (!state.currentUserId) return;

  state.manualWsClose = false;
  if (state.ws) {
    state.ws.close();
  }

  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsURL = `${protocol}://${window.location.host}${API_BASE}/ws?user_id=${encodeURIComponent(state.currentUserId)}`;
  const socket = new WebSocket(wsURL);
  state.ws = socket;

  socket.onopen = () => {
    state.wsConnected = true;
    renderWsBadge();
  };

  socket.onclose = () => {
    state.wsConnected = false;
    renderWsBadge();
    state.ws = null;

    if (!state.manualWsClose && state.currentUserId) {
      if (state.wsTimer) clearTimeout(state.wsTimer);
      state.wsTimer = setTimeout(() => connectWebSocket(), 2200);
    }
  };

  socket.onerror = () => {
    state.wsConnected = false;
    renderWsBadge();
  };

  socket.onmessage = async (event) => {
    try {
      const payload = JSON.parse(event.data);
      await handleSocketEvent(payload);
    } catch (error) {
      toast("WebSocket event parsing xatosi.", "error");
    }
  };
}

async function handleSocketEvent(payload) {
  const type = payload?.type;
  if (!type) return;

  if (type === "new_message") {
    const chatId = Number(payload.chat_id);
    const senderId = Number(payload.sender_id);
    const message = {
      id: Date.now(),
      chatId,
      senderId,
      senderName: payload.sender_name || `User #${senderId}`,
      content: payload.content || "",
      createdAt: payload.created_at || new Date().toISOString(),
      isRead: false,
    };

    if (chatId === state.selectedChatId) {
      state.messages.push(message);
      renderMessages(true);
      await markCurrentChatAsRead(true);
    } else {
      const chat = state.chats.find((item) => item.chatId === chatId);
      if (chat) {
        chat.unreadCount += 1;
        chat.lastMessage = message.content;
        chat.lastMessageAt = message.createdAt;
        renderChatList();
      } else {
        await refreshChats();
      }
      toast(`${message.senderName}: ${clip(message.content, 48)}`, "info");
    }
  }

  if (type === "message_updated") {
    const chatId = Number(payload.chat_id);
    const messageId = Number(payload.message_id);
    if (chatId !== state.selectedChatId) return;

    const message = state.messages.find((item) => item.id === messageId);
    if (!message) return;
    message.content = payload.message_text || message.content;
    renderMessages();
    toast("Xabar tahrirlandi.", "info");
  }

  if (type === "message_deleted") {
    const chatId = Number(payload.chat_id);
    const messageId = Number(payload.message_id);
    if (chatId !== state.selectedChatId) return;

    state.messages = state.messages.filter((item) => item.id !== messageId);
    renderMessages();
    toast("Xabar o'chirildi.", "info");
  }

  if (type === "messages_read") {
    const chatId = Number(payload.chat_id);
    const readerId = Number(payload.reader_id);
    if (chatId !== state.selectedChatId) return;
    if (readerId === state.currentUserId) return;

    let hasUpdates = false;
    state.messages.forEach((message) => {
      if (message.senderId === state.currentUserId && !message.isRead) {
        message.isRead = true;
        hasUpdates = true;
      }
    });
    if (hasUpdates) {
      renderMessages(false);
    }
    toast(`User #${readerId} xabarlarni o'qidi.`, "info");
  }
}

async function loadMemberCandidates(searchTerm) {
  const chat = getSelectedChat();
  if (!chat || chat.chatType !== "group") {
    state.memberCandidates = [];
    renderMemberCandidateList();
    return;
  }

  const query = new URLSearchParams();
  query.set("limit", "20");
  query.set("offset", "0");

  const term = sanitizeUserSearch(searchTerm);
  if (term) {
    query.set("search", term);
  }

  const users = await apiRequest(`/users?${query.toString()}`);
  const memberIDs = new Set(state.members.map((member) => member.id));
  state.memberCandidates = asArray(users)
    .map(normalizeUser)
    .filter((user) => user.id > 0 && !memberIDs.has(user.id));

  renderMemberCandidateList();
}

async function apiRequest(path, options = {}) {
  const method = options.method || "GET";
  const auth = options.auth !== false;
  const body = options.body;

  const headers = {};
  if (auth) {
    if (!state.currentUserId) {
      throw new Error("X-User-ID yuborilmagan.");
    }
    headers["X-User-ID"] = String(state.currentUserId);
  }
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const raw = await response.text();
  let parsed = null;
  if (raw) {
    try {
      parsed = JSON.parse(raw);
    } catch (error) {
      parsed = null;
    }
  }

  if (!response.ok) {
    const message = parsed?.error || `HTTP ${response.status}`;
    throw new Error(message);
  }

  if (parsed && parsed.data !== undefined) {
    return parsed.data;
  }
  return parsed;
}

function renderHealthBadge() {
  const health = state.health;
  if (!health) {
    els.healthBadge.className = "pill neutral";
    els.healthBadge.textContent = "Health: tekshirilmoqda...";
    return;
  }

  const ok = health.status === "available";
  els.healthBadge.className = ok ? "pill ok" : "pill offline";
  els.healthBadge.textContent = `Health: ${health.status || "unknown"} | ${health.ENV || "-"} | ${health.version || "-"}`;
}

function renderWsBadge() {
  if (state.wsConnected) {
    els.wsBadge.className = "pill ok";
    els.wsBadge.textContent = `WS: ulangan (user #${state.currentUserId})`;
  } else {
    els.wsBadge.className = "pill offline";
    els.wsBadge.textContent = "WS: uzilgan";
  }
}

function renderChatList() {
  els.chatList.innerHTML = "";
  els.chatCountBadge.textContent = String(state.chats.length);

  if (!state.chats.length) {
    els.chatList.innerHTML = `<li class="empty">Chatlar topilmadi.</li>`;
    return;
  }

  const fragment = document.createDocumentFragment();
  state.chats.forEach((chat) => {
    const item = document.createElement("li");
    item.className = `list-item chat-item${chat.chatId === state.selectedChatId ? " active" : ""}`;
    item.dataset.chatId = String(chat.chatId);
    item.innerHTML = `
      <div class="chat-item-top">
        <span class="chat-title">${escapeHTML(chat.chatName)}</span>
        <span class="chat-type">${escapeHTML(chat.chatType)}</span>
      </div>
      <div class="chat-last">${escapeHTML(chat.lastMessage || "Xabar yo'q")}</div>
      <div class="chat-foot">
        <span class="chat-time">${formatDate(chat.lastMessageAt || chat.joinedAt)}</span>
        ${chat.unreadCount > 0 ? `<span class="unread">${chat.unreadCount}</span>` : ""}
      </div>
    `;
    fragment.appendChild(item);
  });
  els.chatList.appendChild(fragment);
}

function renderUserList() {
  els.userList.innerHTML = "";

  if (!state.users.length) {
    els.userList.innerHTML = `<li class="empty">Foydalanuvchi topilmadi.</li>`;
    return;
  }

  const fragment = document.createDocumentFragment();
  state.users.forEach((user) => {
    const item = document.createElement("li");
    item.className = "list-item";
    item.innerHTML = `
      <div><strong>${escapeHTML(user.username)}</strong> <span>#${user.id}</span></div>
      <div class="chat-time">${escapeHTML(user.email || "email yo'q")}</div>
      <button type="button" class="btn mini" data-action="start-private" data-user-id="${user.id}">
        <i class="fa-regular fa-message"></i> Chat ochish
      </button>
    `;
    fragment.appendChild(item);
  });
  els.userList.appendChild(fragment);
}

function renderChatMeta() {
  const chat = getSelectedChat();

  if (!chat) {
    els.activeChatName.textContent = "Chat tanlanmagan";
    els.activeChatInfo.textContent = "Chap paneldan chat tanlang yoki yangi chat yarating.";
    els.membersBtn.disabled = true;
    els.editGroupBtn.disabled = true;
    els.markReadBtn.disabled = true;
    els.deleteChatBtn.disabled = true;
    return;
  }

  els.activeChatName.textContent = `${chat.chatName} (#${chat.chatId})`;
  els.activeChatInfo.textContent = `${chat.chatType.toUpperCase()} | role: ${chat.userRole || "-"} | unread: ${chat.unreadCount || 0}`;

  const isGroup = chat.chatType === "group";
  els.membersBtn.disabled = !isGroup;
  els.editGroupBtn.disabled = !isGroup;
  els.markReadBtn.disabled = false;
  els.deleteChatBtn.disabled = false;
}

function renderMessages(shouldScroll = true) {
  els.messagesList.innerHTML = "";

  if (!state.selectedChatId) {
    els.messagesList.innerHTML = `<li class="empty">Xabarlar shu yerda ko'rinadi.</li>`;
    return;
  }

  if (!state.messages.length) {
    els.messagesList.innerHTML = `<li class="empty">Bu chatda xabarlar hali yo'q.</li>`;
    return;
  }

  const fragment = document.createDocumentFragment();
  state.messages.forEach((message) => {
    const mine = message.senderId === state.currentUserId;
    const item = document.createElement("li");
    item.className = `message-item${mine ? " mine" : ""}`;
    item.innerHTML = `
      <div class="message-head">
        <span class="message-author">${escapeHTML(message.senderName || `User #${message.senderId}`)}</span>
        <span class="message-time">
          ${formatDate(message.createdAt)}
          ${mine ? renderMessageStatus(message) : ""}
        </span>
      </div>
      <p class="message-body">${escapeHTML(message.content)}</p>
      ${mine ? `
        <div class="message-tools">
          <button type="button" class="tool-btn edit" data-action="edit-message" data-message-id="${message.id}">
            Tahrirlash
          </button>
          <button type="button" class="tool-btn remove" data-action="delete-message" data-message-id="${message.id}">
            O'chirish
          </button>
        </div>
      ` : ""}
    `;
    fragment.appendChild(item);
  });

  els.messagesList.appendChild(fragment);
  if (shouldScroll) {
    els.messagesList.scrollTop = els.messagesList.scrollHeight;
  }
}
function setMessagesLoading() {
  els.messagesList.innerHTML = `<li class="empty">Xabarlar yuklanmoqda...</li>`;
}

function renderComposerState() {
  const enabled = Boolean(state.selectedChatId && state.currentUserId);
  els.messageInput.disabled = !enabled;
  els.sendMessageBtn.disabled = !enabled;
  if (!enabled) {
    els.messageInput.value = "";
  }
}

function renderMembersList() {
  els.membersList.innerHTML = "";

  if (!state.members.length) {
    els.membersList.innerHTML = `<li class="empty">A'zolar topilmadi.</li>`;
    return;
  }

  const fragment = document.createDocumentFragment();
  state.members.forEach((member) => {
    const canRemove = member.id !== state.currentUserId;
    const item = document.createElement("li");
    item.className = "list-item member-row";
    item.innerHTML = `
      <div>
        <strong>${escapeHTML(member.username)}</strong>
        <div class="chat-time">#${member.id} ${escapeHTML(member.email || "")}</div>
      </div>
      ${canRemove ? `
      <button type="button" class="btn mini danger" data-action="remove-member" data-user-id="${member.id}">
        Chiqarish
      </button>
      ` : `<span class="chat-time">siz</span>`}
    `;
    fragment.appendChild(item);
  });
  els.membersList.appendChild(fragment);
}

function renderMemberCandidateList() {
  els.memberSearchList.innerHTML = "";

  if (!state.memberCandidates.length) {
    els.memberSearchList.innerHTML = `<li class="empty">Qo'shish uchun user topilmadi.</li>`;
    return;
  }

  const fragment = document.createDocumentFragment();
  state.memberCandidates.forEach((user) => {
    const item = document.createElement("li");
    item.className = "list-item member-row";
    item.innerHTML = `
      <div>
        <strong>${escapeHTML(user.username)}</strong>
        <div class="chat-time">#${user.id} ${escapeHTML(user.email || "")}</div>
      </div>
      <button type="button" class="btn mini" data-action="add-member" data-user-id="${user.id}">
        Qo'shish
      </button>
    `;
    fragment.appendChild(item);
  });

  els.memberSearchList.appendChild(fragment);
}

function renderGroupMemberChecklist() {
  els.groupMemberChecklist.innerHTML = "";

  if (!state.groupMemberPool.length) {
    els.groupMemberChecklist.innerHTML = `<div class="empty">Tanlash uchun user yo'q.</div>`;
    updateGroupMemberCount();
    return;
  }

  const fragment = document.createDocumentFragment();
  state.groupMemberPool.forEach((user) => {
    const checked = state.groupSelectedMemberIds.has(user.id);
    const label = document.createElement("label");
    label.className = "check-item";
    label.innerHTML = `
      <input type="checkbox" value="${user.id}" class="group-member-check" ${checked ? "checked" : ""}>
      <span>${escapeHTML(user.username)} (#${user.id})</span>
    `;
    fragment.appendChild(label);
  });
  els.groupMemberChecklist.appendChild(fragment);

  els.groupMemberChecklist.querySelectorAll(".group-member-check").forEach((checkbox) => {
    checkbox.addEventListener("change", () => {
      const userId = Number(checkbox.value);
      if (!userId) return;

      if (checkbox.checked) {
        state.groupSelectedMemberIds.add(userId);
      } else {
        state.groupSelectedMemberIds.delete(userId);
      }
      updateGroupMemberCount();
    });
  });
  updateGroupMemberCount();
}

function getSelectedGroupMembers() {
  return Array.from(state.groupSelectedMemberIds.values())
    .filter((value) => Number.isInteger(value) && value > 0);
}

function updateGroupMemberCount() {
  els.groupMemberCount.textContent = `${state.groupSelectedMemberIds.size} ta tanlangan`;
}

async function loadGroupMemberPool(searchTerm) {
  if (!state.currentUserId) return;

  const query = new URLSearchParams();
  query.set("limit", "20");
  query.set("offset", "0");

  const term = sanitizeUserSearch(searchTerm);
  if (term) {
    query.set("search", term);
  }

  const data = await apiRequest(`/users?${query.toString()}`);
  state.groupMemberPool = asArray(data)
    .map(normalizeUser)
    .filter((user) => user.id > 0 && user.id !== state.currentUserId);

  renderGroupMemberChecklist();
}

function renderEmptyLists() {
  els.chatList.innerHTML = `<li class="empty">Chatlar topilmadi.</li>`;
  els.userList.innerHTML = `<li class="empty">Foydalanuvchi topilmadi.</li>`;
  els.messagesList.innerHTML = `<li class="empty">Xabarlar shu yerda ko'rinadi.</li>`;
}

function openModal(id) {
  const modal = document.getElementById(id);
  if (!modal) return;
  modal.classList.remove("hidden");
}

function closeModal(id) {
  const modal = document.getElementById(id);
  if (!modal) return;
  modal.classList.add("hidden");
}

function ensureSession() {
  if (!state.currentUserId) {
    toast("Avval X-User-ID bilan ulaning.", "error");
    return false;
  }
  return true;
}

function getSelectedChat() {
  return state.chats.find((chat) => chat.chatId === state.selectedChatId) || null;
}

function normalizeChat(raw) {
  return {
    chatId: Number(raw.chat_id ?? raw.chatId ?? raw.ChatID ?? 0),
    chatType: String(raw.chat_type ?? raw.chatType ?? raw.ChatType ?? "private").toLowerCase(),
    chatName: String(raw.chat_name ?? raw.chatName ?? raw.ChatName ?? "No name"),
    userRole: String(raw.user_role ?? raw.userRole ?? raw.UserRole ?? ""),
    joinedAt: raw.joined_at ?? raw.joinedAt ?? raw.JoinedAt ?? "",
    lastMessage: String(raw.last_message ?? raw.lastMessage ?? raw.LastMessage ?? ""),
    lastMessageAt: raw.last_message_at ?? raw.lastMessageAt ?? raw.LastMessageAt ?? "",
    unreadCount: Number(raw.unread_count ?? raw.unreadCount ?? raw.UnreadCount ?? 0),
  };
}

function normalizeUser(raw) {
  return {
    id: Number(raw.id ?? raw.ID ?? 0),
    username: String(raw.username ?? raw.user_name ?? raw.UserName ?? `User #${raw.id || "?"}`),
    email: String(raw.email ?? raw.Email ?? ""),
  };
}

function normalizeMessage(raw, fallbackChatId) {
  return {
    id: Number(raw.id ?? raw.ID ?? Date.now()),
    chatId: Number(raw.chat_id ?? raw.chatId ?? raw.ChatID ?? fallbackChatId ?? state.selectedChatId ?? 0),
    senderId: Number(raw.sender_id ?? raw.senderId ?? raw.SenderID ?? 0),
    senderName: String(raw.sender_name ?? raw.senderName ?? raw.SenderName ?? `User #${raw.sender_id || "?"}`),
    content: String(raw.content ?? raw.message_text ?? raw.messageText ?? raw.MessageText ?? ""),
    createdAt: raw.created_at ?? raw.createdAt ?? raw.CreatedAt ?? new Date().toISOString(),
    isRead: Boolean(raw.is_read ?? raw.isRead ?? raw.IsRead ?? false),
  };
}

function renderMessageStatus(message) {
  const read = Boolean(message?.isRead);
  return `<span class="message-status ${read ? "read" : "sent"}" title="${read ? "O'qilgan" : "Yuborilgan"}">${read ? "&#10003;&#10003;" : "&#10003;"}</span>`;
}

function formatDate(raw) {
  if (!raw) return "-";
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return String(raw);
  return date.toLocaleString("uz-UZ", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function asArray(data) {
  if (Array.isArray(data)) return data;
  return [];
}

function toast(message, type) {
  const item = document.createElement("div");
  item.className = `toast ${type || "info"}`;
  item.textContent = message;
  els.toastStack.appendChild(item);

  setTimeout(() => {
    item.remove();
  }, 3400);
}

function escapeHTML(value) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function debounce(callback, delay) {
  let timer = null;
  return (...args) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => callback(...args), delay);
  };
}

function clip(value, maxLen) {
  const text = String(value || "");
  if (text.length <= maxLen) return text;
  return `${text.slice(0, maxLen)}...`;
}

function sanitizeUserSearch(value) {
  return String(value || "").trim().slice(0, 10);
}
