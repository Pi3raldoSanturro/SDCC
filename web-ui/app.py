from flask import Flask, render_template, request, redirect, session, flash
import subprocess
import json
import sys

sys.stdout.reconfigure(line_buffering=True)


app = Flask(__name__)
app.secret_key = "supersecret"
app.config["SESSION_TYPE"] = "filesystem"

GRPC_SERVICES = {
    "user": "user-service:50051",
    "ticket": "ticket-service:50052",
    "payment": "payment-service:50053",
    "auth": "auth-service:50054"
}

PROTO_PATH = "/app/proto"

def grpc_call(service, method, data, protofile, token=None):
    cmd = [
        "grpcurl",
        "-plaintext",
        "-import-path", PROTO_PATH,             # ✅ Aggiunto
        "-proto", protofile,                    # ✅ file relativo
    ]

    if token:
        cmd += ["-H", f"authorization: Bearer {token}"]

    if data:
        cmd += ["-d", data]

    cmd += [GRPC_SERVICES[service], method]

    print("[DEBUG] Comando grpcurl finale:", cmd, flush=True)
    result = subprocess.run(cmd, capture_output=True, text=True)
    return result.stdout if result.returncode == 0 else result.stderr



@app.route("/")
def index():
    return render_template("index.html")


@app.route("/login", methods=["GET", "POST"])
def login():
    if request.method == "POST":
        username = request.form["username"]
        password = request.form["password"]

        login_data = json.dumps({
            "username": username,
            "password": password
        })

        print("[DEBUG] Eseguo grpc_call con:", login_data)
        output = grpc_call("user", "user.UserService/Login", login_data, "user.proto")
        print("[DEBUG] Risposta da grpc_call:", output)

        try:
            parsed = json.loads(output)
            print("[DEBUG] JSON parsed:", parsed)

            if parsed.get("success") and "token" in parsed:
                session["token"] = parsed["token"]
                session["username"] = username
                session["role"] = parsed["role"]
                session["user_id"] = parsed["userId"]
                print("[DEBUG] Login riuscito. Redirect verso /events")
                return redirect("/events")
            else:
                flash(parsed.get("message", "Login fallito."))
        except json.JSONDecodeError:
            flash("Errore parsing risposta login.")
    return render_template("login.html")

@app.route("/register", methods=["GET", "POST"])
def register():
    if request.method == "POST":
        username = request.form["username"]
        password = request.form["password"]
        role = request.form["role"]

        reg_data = json.dumps({
            "username": username,
            "password": password,
            "role": role
        })

        print("[DEBUG] Eseguo grpc_call di registrazione con:", reg_data)
        output = grpc_call("user", "user.UserService/Register", reg_data, "user.proto")

        try:
            parsed = json.loads(output)
            flash(parsed.get("message", "Registrazione completata."))
            if parsed.get("token"):
                session["token"] = parsed["token"]
                session["username"] = username
                session["role"] = role
                session["user_id"] = parsed["userId"]
                return redirect("/events")
        except:
            flash("Errore durante la registrazione.")
    return render_template("register.html")


@app.route("/events")
def events():
    token = session.get("token")
    if not token:
        return redirect("/login")

    output = grpc_call("ticket", "ticket.TicketService/ListEvents", "{}", "ticket.proto", token)
    print("[DEBUG] Output GetEvents:", output, flush=True)

    try:
        parsed = json.loads(output)
        print("[DEBUG] Parsed JSON:", parsed, flush=True)
        eventi = parsed.get("events", [])
        print("[DEBUG] Lista eventi:", eventi, flush=True)
    except json.JSONDecodeError:
        flash("Errore nel parsing eventi.")
        eventi = []

    return render_template("events.html", eventi=eventi)



@app.route("/buy/<event_id>", methods=["GET", "POST"])
def buy(event_id):
    token = session.get("token")
    if not token:
        return redirect("/login")

    if request.method == "POST":
        quantity = int(request.form.get("quantity", 1))
        user_id = session.get("user_id")
        buy_data = f'{{"eventId": "{event_id}", "quantity": {quantity}, "userId": "{user_id}"}}'

        print("[DEBUG] Comando acquisto grpcurl:", buy_data)
        output = grpc_call("ticket", "ticket.TicketService/PurchaseTicket", buy_data, "ticket.proto", token)
        print("[DEBUG] Output acquisto:", output)

        flash(output)
        return redirect("/events")

    return render_template("buy.html", event_id=event_id)

@app.route("/add-event", methods=["GET", "POST"])
def add_event():
    if session.get("role") != "admin":
        flash("Accesso riservato agli admin.")
        return redirect("/events")

    if request.method == "POST":
        name = request.form.get("name")
        date = request.form.get("date")
        tickets = int(request.form.get("availableTickets"))

        token = session.get("token")
        user_id = session.get("user_id")
        data = json.dumps({
            "userId": user_id,
            "name": name,
            "date": date,
            "availableTickets": tickets
        })

        output = grpc_call("ticket", "ticket.TicketService/AddEvent", data, "ticket.proto", token)
        try:
            parsed = json.loads(output)
            flash(parsed.get("message", "Nessun messaggio restituito."))
        except Exception:
            flash("Errore durante l'aggiunta dell'evento.")
        return redirect("/events")

    return render_template("add_event.html")

@app.route("/delete-event", methods=["GET", "POST"])
def delete_event():
    if session.get("role") != "admin":
        flash("Accesso riservato agli admin.")
        return redirect("/events")

    if request.method == "POST":
        event_id = request.form.get("eventId")
        token = session.get("token")
        user_id = session.get("user_id")
        data = json.dumps({
            "userId": user_id,
            "eventId": event_id
        })

        output = grpc_call("ticket", "ticket.TicketService/DeleteEvent", data, "ticket.proto", token)
        try:
            parsed = json.loads(output)
            flash(parsed.get("message", "Nessun messaggio restituito."))
        except Exception:
            flash("Errore durante la cancellazione dell'evento.")
        return redirect("/events")

    return render_template("delete_event.html")


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080, debug=True)
