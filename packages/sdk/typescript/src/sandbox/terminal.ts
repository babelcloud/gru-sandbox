import { AxiosInstance } from "axios";
import { Language } from "../types/sandbox";



export class TerminalGbox {

    private http: AxiosInstance;
    public sandboxId: string | null;
    constructor(http: AxiosInstance, boxId?: string) {
        this.http = http;
        this.sandboxId = boxId || null;
        const init = async (): Promise<TerminalGbox> => {
            if (boxId) {
                this.sandboxId = boxId;
                return this;
            }else{
                const { data } = await this.http.post('/api/v1/gbox/terminal/create');
                this.sandboxId = data.uid;
                return this;
            }
        };
        return init() as unknown as TerminalGbox;
    }
    async runCode(code: string, language?: Language): Promise<string> {
        if (!language) {
            language = Language.PYTHON;
        }
        if(Language.PYTHON !== language && Language.JAVASCRIPT !== language) {
            throw new Error("Invalid language");
        }
        const { data } = await this.http.post('/api/v1/gbox/terminal/runCode', {
            uid: this.sandboxId,
            code,
            language,
        });
        return data.stdout;
    }

    async runCommand(command: string): Promise<string> {
        const { data } = await this.http.post('/api/v1/gbox/terminal/run', {
            uid: this.sandboxId,
            command,
        });
        return data.stdout;
    }


}