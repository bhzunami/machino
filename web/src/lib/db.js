const dbName = 'machino-todos'
const version = 1

function openDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(dbName, version)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains('kv')) {
        db.createObjectStore('kv')
      }
      if (!db.objectStoreNames.contains('queue')) {
        db.createObjectStore('queue', { keyPath: 'id', autoIncrement: true })
      }
    }
    request.onerror = () => reject(request.error)
    request.onsuccess = () => resolve(request.result)
  })
}

async function transaction(storeName, mode, action) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, mode)
    const store = tx.objectStore(storeName)
    const request = action(store)
    request.onerror = () => reject(request.error)
    request.onsuccess = () => resolve(request.result)
    tx.oncomplete = () => db.close()
    tx.onerror = () => reject(tx.error)
  })
}

export async function getCache(key, fallback) {
  const value = await transaction('kv', 'readonly', (store) => store.get(key)).catch(() => undefined)
  return value ?? fallback
}

export async function setCache(key, value) {
  await transaction('kv', 'readwrite', (store) => store.put(value, key))
}

export async function enqueue(operation) {
  await transaction('queue', 'readwrite', (store) => store.add(operation))
}

export async function getQueue() {
  return transaction('queue', 'readonly', (store) => store.getAll()).catch(() => [])
}

export async function removeQueued(id) {
  await transaction('queue', 'readwrite', (store) => store.delete(id))
}
