// Rendezvous send/recv coordinates tasks (channel() is an alias).

let ch = channel()

let producer = (ch) -> {
    ch.send("a")
    ch.send("b")
    ch.done()
}

let consumer = (ch) -> {
    for true with res = ch.recv(), acc = [] {
        let [msg, done] = res
        if done { break acc }
        acc += [msg]
        res = ch.recv()
    } then acc
}

let results = wait & { producer(ch), consumer(ch) }
results[1]
