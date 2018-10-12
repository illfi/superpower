class LogIn extends React.Component {
    static removeNotification() {
        const notif = document.getElementById("notification");
        notif.hidden = true;
        notif.childNodes[1].textContent = "";
    }

    render() {
        return (
            <section className="hero is-success is-fullheight">
                <div className="hero-body">
                    <div className="container has-text-centered">
                        <div className="column is-4 is-offset-4">
                            <div className="box">
                                <form method="post" id="login-form" onSubmit={this.login.bind(this)}>
                                    <div style={{display: "grid"}}>
                                        <Logo/>
                                        <div style={{marginBottom: '15px'}} hidden id="notification"
                                             className="notification is-danger">
                                            <button id="closeNotification" type="button" className="delete"
                                                    onClick={LogIn.removeNotification}/>
                                        </div>
                                    </div>
                                    <div className="field">
                                        <div className="control">
                                            <input name="email" className="input" type="email" placeholder="email"
                                                   autoFocus=""/>
                                        </div>
                                    </div>
                                    <div className="field">
                                        <div className="control">
                                            <input name="pw_hash" className="input" type="password"
                                                   placeholder="password" autoFocus=""/>
                                        </div>
                                    </div>
                                    <button type="submit" className="button is-block is-info is-fullwidth">login
                                    </button>
                                </form>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        )
    }

    login(a) {
        a.preventDefault();
        const form = new FormData(document.querySelector("#login-form"));
        const email = form.get("email");
        const pass = form.get("password");
        // hash our password
        // cum...
        // login
        axios.post('/api/auth/login', form).then(d => {
            if (d.response.data.error) {
                a.preventDefault();
            }
        }).catch(e => {
            const notif = document.getElementById("notification");
            if (notif.hidden && notif.childNodes.length === 1) {
                console.log(e.response.data);
                const node = document.createTextNode(e.response.data.response);
                notif.appendChild(node);
            } else {
                notif.childNodes[1].textContent = e.response.data.response;
            }
            notif.hidden = false;
            return false;
        });
        this.props.updateAuth();
        return false;
    }
}

const Logo = () => (
    <div style={{float: 'left', display: 'inline-flex', marginBottom: '15px'}}>
        <div style={{
            backgroundColor: '#e24040',
            width: '8px',
            height: '60px',
            marginRight: '5px',
            marginTop: '5px'
        }}/>
        <div style={{backgroundColor: '#141414', width: '8px', height: '65px', marginRight: '5px'}}/>
        <div style={{backgroundColor: '#141414', width: '8px', height: '65px'}}/>

    </div>
);

class App extends React.Component {

    constructor() {
        super();
        this.state = {
            authenticated: false
        };
    }

    updateAuth() {
        this.setState((prev, ps) => {
            let res = fetch("/api/auth/check")
                .then(res => res.json())
                .then(dat => {
                    return {authenticated: dat.response}
                })
        });
    }

    /*componentWillMount() {
        fetch("/api/auth/check")
            .then(res => res.json())
            .then(dat => this.setState({ authenticated: dat.response }))
    }*/

    render() {
        if (this.state.authenticated) {
            return (<Home/>)
        } else {
            return (<LogIn updateAuth={this.updateAuth}/>)
        }
    }

}

ReactDOM.render(<App/>, document.getElementById('app'));
