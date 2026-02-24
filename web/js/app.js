const API_BASE = "/api/v1";
const TOKEN_STORAGE_KEY = "chatx_access_token";
const LAST_EMAIL_STORAGE_KEY = "chatx_last_email";

const state = {
  token: null,
  currentUserId: null,
  currentUsername: null,
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
  authScreen: document.getElementById("authScreen"),
  appShell: document.getElementById("appShell"),
  authHint: document.getElementById("authHint"),
  loginTabBtn: document.getElementById("loginTabBtn"),
  registerTabBtn: document.getElementById("registerTabBtn"),
  loginPane: document.getElementById("loginPane"),
  registerPane: document.getElementById("registerPane"),
  loginForm: document.getElementById("loginForm"),
  loginEmailInput: document.getElementById("loginEmailInput"),
  loginPasswordInput: document.getElementById("loginPasswordInput"),
  registerForm: document.getElementById("registerForm"),
  registerUsernameInput: document.getElementById("registerUsernameInput"),
  registerEmailInput: document.getElementById("registerEmailInput"),
  registerPasswordInput: document.getElementById("registerPasswordInput"),
  registerConfirmInput: document.getElementById("registerConfirmInput"),
  activateForm: document.getElementById("activateForm"),
  activationTokenInput: document.getElementById("activationTokenInput"),
  healthBadge: document.getElementById("healthBadge"),
  wsBadge: document.getElementById("wsBadge"),
  sessionUserBadge: document.getElementById("sessionUserBadge"),
  logoutBtn: document.getElementById("logoutBtn"),
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
  privateForm: document.getElementById("privateForm"),
  privateReceiverSelect: document.getElementById("privateReceiverSelect"),
  groupForm: document.getElementById("groupForm"),
  groupNameInput: document.getElementById("groupNameInput"),
  groupDescInput: document.getElementById("groupDescInput"),
  groupMemberSearchInput: document.getElementById("groupMemberSearchInput"),
  groupMemberChecklist: document.getElementById("groupMemberChecklist"),
  groupMemberCount: document.getElementById("groupMemberCount"),
  editGroupForm: document.getElementById("editGroupForm"),
  editGroupNameInput: document.getElementById("editGroupNameInput"),
  editGroupDescInput: document.getElementById("editGroupDescInput"),
  memberSearchInput: document.getElementById("memberSearchInput"),
  memberSearchList: document.getElementById("memberSearchList"),
  membersList: document.getElementById("membersList"),
  toastStack: document.getElementById("toastStack"),
};

document.addEventListener("DOMContentLoaded", init);

function init() {
  bindEvents();
  setAuthView("login");
  renderHealthBadge();
  renderWsBadge();
  renderSessionBadge();
  renderEmptyLists();
  renderChatMeta();
  renderComposerState();

  const savedEmail = localStorage.getItem(LAST_EMAIL_STORAGE_KEY);
  if (savedEmail) els.loginEmailInput.value = savedEmail;

  hydrateSessionFromStorage();
  fetchHealth().catch(() => {});
  setInterval(() => fetchHealth().catch(() => {}), 30000);
}

function bindEvents() {
  els.loginTabBtn.addEventListener("click", () => setAuthView("login"));
  els.registerTabBtn.addEventListener("click", () => setAuthView("register"));
  els.loginForm.addEventListener("submit", submitLogin);
  els.registerForm.addEventListener("submit", submitRegister);
  els.activateForm.addEventListener("submit", submitActivation);

  els.logoutBtn.addEventListener("click", () => logoutSession(true));
  els.refreshAllBtn.addEventListener("click", refreshAllData);
  els.chatSearchInput.addEventListener("input", debounce(() => refreshChats().catch(showError), 260));
  els.userSearchInput.addEventListener("input", debounce(() => refreshUsers().catch(showError), 260));
  els.groupMemberSearchInput.addEventListener("input", debounce(() => loadGroupMemberPool(els.groupMemberSearchInput.value.trim()).catch(showError), 260));
  els.memberSearchInput.addEventListener("input", debounce(() => loadMemberCandidates(els.memberSearchInput.value.trim()).catch(showError), 260));

  els.newPrivateBtn.addEventListener("click", openPrivateModal);
  els.newGroupBtn.addEventListener("click", openGroupModal);
  els.membersBtn.addEventListener("click", openMembersModal);
  els.editGroupBtn.addEventListener("click", openEditGroupModal);
  els.markReadBtn.addEventListener("click", () => markCurrentChatAsRead(false));
  els.deleteChatBtn.addEventListener("click", deleteCurrentChat);

  els.privateForm.addEventListener("submit", submitPrivateChat);
  els.groupForm.addEventListener("submit", submitGroupChat);
  els.editGroupForm.addEventListener("submit", submitGroupUpdate);
  els.composerForm.addEventListener("submit", submitMessage);

  els.chatList.addEventListener("click", (event) => {
    const item = event.target.closest("[data-chat-id]");
    if (!item) return;
    const chatID = Number(item.dataset.chatId);
    if (chatID > 0) selectChat(chatID);
  });

  els.userList.addEventListener("click", (event) => {
    const button = event.target.closest("[data-action='start-private']");
    if (!button) return;
    const userID = Number(button.dataset.userId);
    if (userID > 0) createPrivateChat(userID);
  });

  els.messagesList.addEventListener("click", (event) => {
    const button = event.target.closest("[data-action]");
    if (!button) return;
    const messageID = Number(button.dataset.messageId);
    if (!messageID) return;
    if (button.dataset.action === "edit-message") editMessage(messageID);
    if (button.dataset.action === "delete-message") deleteMessage(messageID);
  });

  els.membersList.addEventListener("click", (event) => {
    const button = event.target.closest("[data-action='remove-member']");
    if (!button) return;
    const userID = Number(button.dataset.userId);
    if (userID > 0) removeMember(userID);
  });

  els.memberSearchList.addEventListener("click", (event) => {
    const button = event.target.closest("[data-action='add-member']");
    if (!button) return;
    const userID = Number(button.dataset.userId);
    if (userID > 0) addMember(userID);
  });

  document.querySelectorAll("[data-close-modal]").forEach((button) => {
    button.addEventListener("click", () => closeModal(button.dataset.closeModal));
  });

  document.querySelectorAll(".modal").forEach((modal) => {
    modal.addEventListener("click", (event) => {
      if (event.target === modal) modal.classList.add("hidden");
    });
  });
}

function hydrateSessionFromStorage() {
  const token = localStorage.getItem(TOKEN_STORAGE_KEY);
  if (!token || !setSessionToken(token, false)) {
    localStorage.removeItem(TOKEN_STORAGE_KEY);
    showAuthScreen();
    return;
  }

  showWorkspace();
  refreshAllData().catch(showError);
  connectWebSocket();
}

function setSessionToken(token, persist) {
  const claims = parseTokenClaims(token);
  const userID = Number(claims?.sub || 0);
  const username = normalizeUsername(claims?.username);
  const exp = Number(claims?.exp || 0);
  if (!Number.isInteger(userID) || userID <= 0) return false;
  if (exp > 0 && Date.now() >= exp * 1000) return false;

  state.token = token;
  state.currentUserId = userID;
  state.currentUsername = username || null;
  if (persist) localStorage.setItem(TOKEN_STORAGE_KEY, token);
  renderSessionBadge();
  renderWsBadge();
  return true;
}

function clearSessionState() {
  state.token = null;
  state.currentUserId = null;
  state.currentUsername = null;
  state.chats = [];
  state.users = [];
  state.groupMemberPool = [];
  state.groupSelectedMemberIds = new Set();
  state.selectedChatId = null;
  state.messages = [];
  state.members = [];
  state.memberCandidates = [];
  disconnectWebSocket();
  renderSessionBadge();
  renderWsBadge();
  renderChatList();
  renderUserList();
  renderMessages();
  renderChatMeta();
  renderComposerState();
}
function logoutSession(showToastMessage) {
  clearSessionState();
  localStorage.removeItem(TOKEN_STORAGE_KEY);
  showAuthScreen();
  if (showToastMessage) toast("Sessiya yakunlandi.", "info");
}

function showWorkspace() {
  els.authScreen.classList.add("hidden");
  els.appShell.classList.remove("hidden");
}

function showAuthScreen() {
  els.appShell.classList.add("hidden");
  els.authScreen.classList.remove("hidden");
}

function setAuthView(view) {
  const loginActive = view !== "register";
  els.loginTabBtn.classList.toggle("active", loginActive);
  els.registerTabBtn.classList.toggle("active", !loginActive);
  els.loginPane.classList.toggle("active", loginActive);
  els.registerPane.classList.toggle("active", !loginActive);
}

function setAuthHint(message, type) {
  els.authHint.className = `auth-hint ${type || "info"}`;
  els.authHint.textContent = message;
}

async function submitLogin(event) {
  event.preventDefault();
  const email = els.loginEmailInput.value.trim();
  const password = els.loginPasswordInput.value;
  if (!email || !password) {
    setAuthHint("Email va parolni kiriting.", "error");
    return;
  }

  const submit = event.submitter;
  if (submit) submit.disabled = true;

  try {
    const data = await apiRequest("/users/authentication/token", {
      method: "POST",
      auth: false,
      body: { email, password },
    });

    const token = typeof data === "string" ? data : data?.token;
    if (!token || !setSessionToken(token, true)) {
      throw new Error("Yaroqsiz token qaytdi.");
    }

    localStorage.setItem(LAST_EMAIL_STORAGE_KEY, email);
    els.loginPasswordInput.value = "";
    setAuthHint("Muvaffaqiyatli login qilindi.", "ok");
    showWorkspace();
    toast(`Xush kelibsiz, ${getCurrentUserDisplayName()}.`, "ok");
    await refreshAllData();
    connectWebSocket();
  } catch (error) {
    setAuthHint(error.message, "error");
    toast(error.message, "error");
  } finally {
    if (submit) submit.disabled = false;
  }
}

async function submitRegister(event) {
  event.preventDefault();
  const username = els.registerUsernameInput.value.trim();
  const email = els.registerEmailInput.value.trim();
  const password = els.registerPasswordInput.value;
  const confirm = els.registerConfirmInput.value;

  if (!username || !email || !password || !confirm) {
    setAuthHint("Barcha maydonlarni to'ldiring.", "error");
    return;
  }
  if (password !== confirm) {
    setAuthHint("Parollar mos emas.", "error");
    return;
  }

  const submit = event.submitter;
  if (submit) submit.disabled = true;

  try {
    const data = await apiRequest("/users/authentication", {
      method: "POST",
      auth: false,
      body: { username, email, password },
    });

    setAuthHint(data?.message || "Registratsiya muvaffaqiyatli. Emaildagi token bilan accountni aktivatsiya qiling.", "ok");
    toast("Registratsiya muvaffaqiyatli.", "ok");
    els.registerUsernameInput.value = "";
    els.registerEmailInput.value = "";
    els.registerPasswordInput.value = "";
    els.registerConfirmInput.value = "";
    els.loginEmailInput.value = email;
    setAuthView("login");
  } catch (error) {
    setAuthHint(error.message, "error");
    toast(error.message, "error");
  } finally {
    if (submit) submit.disabled = false;
  }
}

async function submitActivation(event) {
  event.preventDefault();
  const token = els.activationTokenInput.value.trim();
  if (!token) {
    setAuthHint("Aktivatsiya tokenini kiriting.", "error");
    return;
  }

  try {
    const data = await apiRequest(`/users/activate/${encodeURIComponent(token)}`, {
      method: "PUT",
      auth: false,
    });
    setAuthHint(data?.message || "Account aktivatsiya qilindi.", "ok");
    toast("Account aktivatsiya qilindi.", "ok");
    els.activationTokenInput.value = "";
    setAuthView("login");
  } catch (error) {
    setAuthHint(error.message, "error");
    toast(error.message, "error");
  }
}

async function refreshAllData() {
  if (!ensureSession()) {
    await fetchHealth();
    return;
  }

  const [healthRes, usersRes, chatsRes] = await Promise.allSettled([
    fetchHealth(),
    refreshUsers(),
    refreshChats(),
  ]);

  if (healthRes.status === "rejected") toast(healthRes.reason.message, "error");
  if (usersRes.status === "rejected") toast(usersRes.reason.message, "error");
  if (chatsRes.status === "rejected") toast(chatsRes.reason.message, "error");

  if (state.selectedChatId && !state.chats.some((chat) => chat.chatId === state.selectedChatId)) {
    state.selectedChatId = null;
    state.messages = [];
    renderMessages();
    renderChatMeta();
    renderComposerState();
  }
}

async function fetchHealth() {
  state.health = await apiRequest("/health", { auth: false });
  renderHealthBadge();
}

async function refreshUsers() {
  if (!ensureSession()) return;
  const query = new URLSearchParams({ limit: "20", offset: "0" });
  const search = sanitizeUserSearch(els.userSearchInput.value.trim());
  if (search) query.set("search", search);
  const data = await apiRequest(`/users?${query.toString()}`);
  state.users = asArray(data).map(normalizeUser).filter((user) => user.id > 0);
  renderUserList();
}

async function refreshChats() {
  if (!ensureSession()) return;
  const query = new URLSearchParams();
  const search = els.chatSearchInput.value.trim();
  if (search) query.set("search", search);
  const path = query.toString() ? `/chats?${query.toString()}` : "/chats";
  const data = await apiRequest(path);
  state.chats = asArray(data).map(normalizeChat).filter((chat) => chat.chatId > 0);
  renderChatList();
}

async function selectChat(chatID) {
  if (!ensureSession()) return;
  state.selectedChatId = chatID;
  renderChatList();
  renderChatMeta();
  renderComposerState();
  setMessagesLoading();

  try {
    const rawMessages = await apiRequest(`/chats/${chatID}/messages`);
    state.messages = asArray(rawMessages).map((message) => normalizeMessage(message, chatID));
    renderMessages();
    markCurrentChatAsRead(true).catch(() => {});
  } catch (error) {
    state.messages = [];
    renderMessages();
    toast(error.message, "error");
  }
}

async function markCurrentChatAsRead(silent) {
  if (!ensureSession() || !state.selectedChatId) return;
  await apiRequest(`/messages/chats/${state.selectedChatId}/read`, { method: "PATCH" });
  const chat = getSelectedChat();
  if (chat) chat.unreadCount = 0;
  renderChatList();
  if (!silent) toast("Chat read holatiga o'tkazildi.", "ok");
}

async function submitMessage(event) {
  event.preventDefault();
  if (!ensureSession()) return;
  if (!state.selectedChatId) return toast("Chat tanlanmagan.", "error");

  const text = els.messageInput.value.trim();
  if (!text) return;

  try {
    const created = await apiRequest("/messages", {
      method: "POST",
      body: { chat_id: state.selectedChatId, message_text: text },
    });
    state.messages.push(normalizeMessage(created, state.selectedChatId));
    els.messageInput.value = "";
    renderMessages(true);
    await refreshChats();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function editMessage(messageID) {
  const message = state.messages.find((item) => item.id === messageID);
  if (!message) return;
  if (message.senderId !== state.currentUserId) return toast("Faqat o'zingiz yuborgan xabarni tahrirlaysiz.", "error");

  const nextText = prompt("Yangi xabar matni:", message.content);
  if (nextText === null) return;
  const text = nextText.trim();
  if (!text || text === message.content) return;

  try {
    await apiRequest(`/messages/${messageID}`, { method: "PATCH", body: { message_text: text } });
    message.content = text;
    renderMessages();
    toast("Xabar yangilandi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function deleteMessage(messageID) {
  if (!confirm("Xabarni o'chirishni tasdiqlaysizmi?")) return;
  try {
    await apiRequest(`/messages/${messageID}`, { method: "DELETE" });
    state.messages = state.messages.filter((item) => item.id !== messageID);
    renderMessages();
    await refreshChats();
    toast("Xabar o'chirildi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}
async function createPrivateChat(receiverID) {
  if (!ensureSession()) return;
  try {
    const response = await apiRequest("/chats", { method: "POST", body: { receiver_id: receiverID } });
    const chatID = Number(response?.chat_id || response?.chatId || response);
    await refreshChats();
    if (chatID > 0) await selectChat(chatID);
    closeModal("privateModal");
    toast("Private chat yaratildi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

function openPrivateModal() {
  if (!ensureSession()) return;
  els.privateReceiverSelect.innerHTML = "";
  const users = state.users.filter((user) => user.id !== state.currentUserId);
  if (!users.length) {
    const option = document.createElement("option");
    option.textContent = "Foydalanuvchi topilmadi";
    option.value = "";
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

async function submitPrivateChat(event) {
  event.preventDefault();
  const receiverID = Number(els.privateReceiverSelect.value);
  if (!receiverID) return toast("Receiver tanlang.", "error");
  await createPrivateChat(receiverID);
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
  await loadGroupMemberPool("").catch(showError);
}

async function submitGroupChat(event) {
  event.preventDefault();
  const name = els.groupNameInput.value.trim();
  const description = els.groupDescInput.value.trim();
  const memberIDs = getSelectedGroupMembers();
  if (!name) return toast("Group nomi bo'sh bo'lmasin.", "error");
  if (!memberIDs.length) return toast("Kamida bitta a'zo tanlang.", "error");

  try {
    const response = await apiRequest("/groups", { method: "POST", body: { name, description, member_ids: memberIDs } });
    const chatID = Number(response?.chat_id || response?.chatId || response);
    await refreshChats();
    if (chatID > 0) await selectChat(chatID);
    closeModal("groupModal");
    toast("Group yaratildi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

function openEditGroupModal() {
  const chat = getSelectedChat();
  if (!chat) return toast("Avval chat tanlang.", "error");
  if (chat.chatType !== "group") return toast("Bu amal faqat group chat uchun.", "error");
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
  if (!name) return toast("Group nomi bo'sh bo'lmasin.", "error");

  try {
    await apiRequest(`/groups/${chat.chatId}`, { method: "PATCH", body: { name, description } });
    await refreshChats();
    renderChatMeta();
    closeModal("editGroupModal");
    toast("Group yangilandi.", "ok");
  } catch (error) {
    toast(error.message, "error");
  }
}

async function deleteCurrentChat() {
  const chat = getSelectedChat();
  if (!chat) return toast("Avval chat tanlang.", "error");
  if (!confirm(`"${chat.chatName}" chatini o'chirishni tasdiqlaysizmi?`)) return;

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
  if (!chat) return toast("Avval chat tanlang.", "error");
  if (chat.chatType !== "group") return toast("A'zolar ro'yxati faqat group chatda mavjud.", "error");

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

async function addMember(userID) {
  const chat = getSelectedChat();
  if (!chat) return;

  try {
    await apiRequest(`/groups/${chat.chatId}/members`, { method: "POST", body: { user_id: userID } });
    toast("A'zo groupga qo'shildi.", "ok");
    await openMembersModal();
    await refreshChats();
  } catch (error) {
    toast(error.message, "error");
  }
}

async function removeMember(userID) {
  const chat = getSelectedChat();
  if (!chat) return;
  const member = state.members.find((item) => item.id === userID);
  const title = member ? member.username : `#${userID}`;
  if (!confirm(`${title} ni groupdan chiqarishni tasdiqlaysizmi?`)) return;

  try {
    await apiRequest(`/groups/${chat.chatId}/${userID}/member`, { method: "DELETE" });
    toast("A'zo chiqarildi.", "ok");
    await openMembersModal();
    await refreshChats();
    if (userID === state.currentUserId) {
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
  if (!state.token) return;
  state.manualWsClose = false;
  if (state.ws) state.ws.close();

  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const wsURL = `${protocol}://${window.location.host}${API_BASE}/ws?token=${encodeURIComponent(state.token)}`;
  state.ws = new WebSocket(wsURL);

  state.ws.onopen = () => {
    state.wsConnected = true;
    renderWsBadge();
  };

  state.ws.onclose = () => {
    state.wsConnected = false;
    renderWsBadge();
    state.ws = null;
    if (!state.manualWsClose && state.token) {
      if (state.wsTimer) clearTimeout(state.wsTimer);
      state.wsTimer = setTimeout(connectWebSocket, 2200);
    }
  };

  state.ws.onerror = () => {
    state.wsConnected = false;
    renderWsBadge();
  };

  state.ws.onmessage = async (event) => {
    try {
      await handleSocketEvent(JSON.parse(event.data));
    } catch {
      toast("WebSocket event parsing xatosi.", "error");
    }
  };
}

function disconnectWebSocket() {
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
}

async function handleSocketEvent(payload) {
  const type = payload?.type;
  if (!type) return;

  if (type === "new_message") {
    const chatID = Number(payload.chat_id);
    const senderID = Number(payload.sender_id);
    const message = {
      id: Date.now(),
      chatId: chatID,
      senderId: senderID,
      senderName: normalizeUsername(payload.sender_name) || getUserDisplayName(senderID),
      content: payload.content || "",
      createdAt: payload.created_at || new Date().toISOString(),
      isRead: false,
    };

    if (chatID === state.selectedChatId) {
      state.messages.push(message);
      renderMessages(true);
      await markCurrentChatAsRead(true);
    } else {
      const chat = state.chats.find((item) => item.chatId === chatID);
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
    const chatID = Number(payload.chat_id);
    const messageID = Number(payload.message_id);
    if (chatID !== state.selectedChatId) return;
    const message = state.messages.find((item) => item.id === messageID);
    if (!message) return;
    message.content = payload.message_text || message.content;
    renderMessages();
    toast("Xabar tahrirlandi.", "info");
  }

  if (type === "message_deleted") {
    const chatID = Number(payload.chat_id);
    const messageID = Number(payload.message_id);
    if (chatID !== state.selectedChatId) return;
    state.messages = state.messages.filter((item) => item.id !== messageID);
    renderMessages();
    toast("Xabar o'chirildi.", "info");
  }

  if (type === "messages_read") {
    const chatID = Number(payload.chat_id);
    const readerID = Number(payload.reader_id);
    if (chatID !== state.selectedChatId || readerID === state.currentUserId) return;
    let hasUpdates = false;
    state.messages.forEach((message) => {
      if (message.senderId === state.currentUserId && !message.isRead) {
        message.isRead = true;
        hasUpdates = true;
      }
    });
    if (hasUpdates) renderMessages(false);
    toast(`${getUserDisplayName(readerID)} xabarlarni o'qidi.`, "info");
  }
}

async function loadMemberCandidates(searchTerm) {
  const chat = getSelectedChat();
  if (!chat || chat.chatType !== "group") {
    state.memberCandidates = [];
    renderMemberCandidateList();
    return;
  }

  const query = new URLSearchParams({ limit: "20", offset: "0" });
  const term = sanitizeUserSearch(searchTerm);
  if (term) query.set("search", term);

  const users = await apiRequest(`/users?${query.toString()}`);
  const memberIDs = new Set(state.members.map((member) => member.id));
  state.memberCandidates = asArray(users).map(normalizeUser).filter((user) => user.id > 0 && !memberIDs.has(user.id));
  renderMemberCandidateList();
}

async function loadGroupMemberPool(searchTerm) {
  if (!ensureSession()) return;
  const query = new URLSearchParams({ limit: "20", offset: "0" });
  const term = sanitizeUserSearch(searchTerm);
  if (term) query.set("search", term);

  const data = await apiRequest(`/users?${query.toString()}`);
  state.groupMemberPool = asArray(data).map(normalizeUser).filter((user) => user.id > 0 && user.id !== state.currentUserId);
  renderGroupMemberChecklist();
}

async function apiRequest(path, options = {}) {
  const method = options.method || "GET";
  const auth = options.auth !== false;
  const body = options.body;

  const headers = {};
  if (auth) {
    if (!state.token) throw new Error("Avval login qiling.");
    headers.Authorization = `Bearer ${state.token}`;
  }
  if (body !== undefined) headers["Content-Type"] = "application/json";

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
    } catch {
      parsed = null;
    }
  }

  if (!response.ok) {
    const message = parsed?.error || `HTTP ${response.status}`;
    if (response.status === 401 && auth) {
      logoutSession(false);
      setAuthHint("Sessiya muddati tugadi. Qayta login qiling.", "error");
    }
    throw new Error(message);
  }

  if (parsed && parsed.data !== undefined) return parsed.data;
  return parsed;
}

function renderHealthBadge() {
  if (!state.health) {
    els.healthBadge.className = "pill neutral";
    els.healthBadge.textContent = "Health: tekshirilmoqda...";
    return;
  }
  const ok = state.health.status === "available";
  els.healthBadge.className = ok ? "pill ok" : "pill offline";
  els.healthBadge.textContent = `Health: ${state.health.status || "unknown"} | ${state.health.ENV || "-"} | ${state.health.version || "-"}`;
}

function renderWsBadge() {
  if (!state.token) {
    els.wsBadge.className = "pill neutral";
    els.wsBadge.textContent = "WS: login kutilmoqda";
    return;
  }

  if (state.wsConnected) {
    els.wsBadge.className = "pill ok";
    els.wsBadge.textContent = `WS: ulangan (${getCurrentUserDisplayName()})`;
  } else {
    els.wsBadge.className = "pill offline";
    els.wsBadge.textContent = "WS: uzilgan";
  }
}

function renderSessionBadge() {
  els.sessionUserBadge.textContent = state.currentUserId ? `Ulangan: ${getCurrentUserDisplayName()}` : "User aniqlanmagan";
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
        <span class="message-author">${escapeHTML(message.senderName || getUserDisplayName(message.senderId))}</span>
        <span class="message-time">
          ${formatDate(message.createdAt)}
          ${mine ? renderMessageStatus(message) : ""}
        </span>
      </div>
      <p class="message-body">${escapeHTML(message.content)}</p>
      ${mine ? `
        <div class="message-tools">
          <button type="button" class="tool-btn edit" data-action="edit-message" data-message-id="${message.id}">Tahrirlash</button>
          <button type="button" class="tool-btn remove" data-action="delete-message" data-message-id="${message.id}">O'chirish</button>
        </div>
      ` : ""}
    `;
    fragment.appendChild(item);
  });

  els.messagesList.appendChild(fragment);
  if (shouldScroll) els.messagesList.scrollTop = els.messagesList.scrollHeight;
}

function setMessagesLoading() {
  els.messagesList.innerHTML = `<li class="empty">Xabarlar yuklanmoqda...</li>`;
}

function renderComposerState() {
  const enabled = Boolean(state.selectedChatId && state.currentUserId && state.token);
  els.messageInput.disabled = !enabled;
  els.sendMessageBtn.disabled = !enabled;
  if (!enabled) els.messageInput.value = "";
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
      ${canRemove ? `<button type="button" class="btn mini danger" data-action="remove-member" data-user-id="${member.id}">Chiqarish</button>` : `<span class="chat-time">siz</span>`}
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
      <button type="button" class="btn mini" data-action="add-member" data-user-id="${user.id}">Qo'shish</button>
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
    label.innerHTML = `<input type="checkbox" value="${user.id}" class="group-member-check" ${checked ? "checked" : ""}><span>${escapeHTML(user.username)} (#${user.id})</span>`;
    fragment.appendChild(label);
  });
  els.groupMemberChecklist.appendChild(fragment);

  els.groupMemberChecklist.querySelectorAll(".group-member-check").forEach((checkbox) => {
    checkbox.addEventListener("change", () => {
      const userID = Number(checkbox.value);
      if (!userID) return;
      if (checkbox.checked) state.groupSelectedMemberIds.add(userID);
      else state.groupSelectedMemberIds.delete(userID);
      updateGroupMemberCount();
    });
  });

  updateGroupMemberCount();
}

function getSelectedGroupMembers() {
  return Array.from(state.groupSelectedMemberIds).filter((id) => Number.isInteger(id) && id > 0);
}

function updateGroupMemberCount() {
  els.groupMemberCount.textContent = `${state.groupSelectedMemberIds.size} ta tanlangan`;
}

function renderEmptyLists() {
  els.chatList.innerHTML = `<li class="empty">Chatlar topilmadi.</li>`;
  els.userList.innerHTML = `<li class="empty">Foydalanuvchi topilmadi.</li>`;
  els.messagesList.innerHTML = `<li class="empty">Xabarlar shu yerda ko'rinadi.</li>`;
}

function openModal(id) {
  const modal = document.getElementById(id);
  if (modal) modal.classList.remove("hidden");
}

function closeModal(id) {
  const modal = document.getElementById(id);
  if (modal) modal.classList.add("hidden");
}

function ensureSession() {
  return Boolean(state.token && state.currentUserId);
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

function normalizeMessage(raw, fallbackChatID) {
  const senderID = Number(raw.sender_id ?? raw.senderId ?? raw.SenderID ?? 0);
  return {
    id: Number(raw.id ?? raw.ID ?? Date.now()),
    chatId: Number(raw.chat_id ?? raw.chatId ?? raw.ChatID ?? fallbackChatID ?? state.selectedChatId ?? 0),
    senderId: senderID,
    senderName: normalizeUsername(raw.sender_name ?? raw.senderName ?? raw.SenderName) || getUserDisplayName(senderID),
    content: String(raw.content ?? raw.message_text ?? raw.messageText ?? raw.MessageText ?? ""),
    createdAt: raw.created_at ?? raw.createdAt ?? raw.CreatedAt ?? new Date().toISOString(),
    isRead: Boolean(raw.is_read ?? raw.isRead ?? raw.IsRead ?? false),
  };
}
function renderMessageStatus(message) {
  const read = Boolean(message?.isRead);
  return `<span class="message-status ${read ? "read" : "sent"}" title="${read ? "O'qilgan" : "Yuborilgan"}">${read ? "&#10003;&#10003;" : "&#10003;"}</span>`;
}

function parseTokenClaims(token) {
  const parts = String(token || "").split(".");
  if (parts.length !== 3) return null;

  const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
  const padded = payload.padEnd(Math.ceil(payload.length / 4) * 4, "=");

  try {
    const decoded = atob(padded);
    const bytes = Uint8Array.from(decoded, (ch) => ch.charCodeAt(0));
    const json = new TextDecoder().decode(bytes);
    return JSON.parse(json);
  } catch {
    return null;
  }
}

function normalizeUsername(value) {
  const text = String(value ?? "").trim();
  return text || null;
}

function getCurrentUserDisplayName() {
  return state.currentUsername || `User #${state.currentUserId}`;
}

function getUserDisplayName(userID) {
  if (!Number.isInteger(userID) || userID <= 0) {
    return "Unknown user";
  }

  if (state.currentUserId === userID && state.currentUsername) {
    return state.currentUsername;
  }

  const pools = [state.users, state.members, state.groupMemberPool, state.memberCandidates];
  for (const pool of pools) {
    const user = pool.find((item) => item.id === userID && normalizeUsername(item.username));
    if (user) return user.username;
  }

  const fromMessage = state.messages.find((item) => item.senderId === userID && normalizeUsername(item.senderName));
  if (fromMessage) return fromMessage.senderName;

  return `User #${userID}`;
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
  return Array.isArray(data) ? data : [];
}

function showError(error) {
  toast(error?.message || "Noma'lum xatolik", "error");
}

function toast(message, type) {
  const item = document.createElement("div");
  item.className = `toast ${type || "info"}`;
  item.textContent = message;
  els.toastStack.appendChild(item);
  setTimeout(() => item.remove(), 3400);
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
  return text.length <= maxLen ? text : `${text.slice(0, maxLen)}...`;
}

function sanitizeUserSearch(value) {
  return String(value || "").trim().slice(0, 10);
}
